// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package persist provides a persistence layer for the in-memory cache
package persist

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

var mysqlSchema = `
CREATE TABLE IF NOT EXISTS persist (
	id INT AUTO_INCREMENT PRIMARY KEY,
	saved TIMESTAMP DEFAULT '1970-01-01 00:00:01',
	k VARCHAR(255) NOT NULL,
	v MEDIUMBLOB,
	UNIQUE KEY unique_k (k),
	INDEX saved_idx (saved)
);`

// sqlItem maps to schema
type sqlItem struct {
	ID    int64     `db:"id"`
	Saved time.Time `db:"saved"`
	Key   string    `db:"k"`
	Value []byte    `db:"v"`
}

type MySQL struct {
	cache *cache.Cache
	db    *sqlx.DB
	path  string
}

// NewMySQL returns a new MySQL cache
func NewMySQL(cfg Config) (*MySQL, error) {
	dbx, err := sqlx.Connect("mysql", cfg.Path+"?parseTime=true")
	if err != nil {
		return nil, err
	}

	m := &MySQL{
		db:   dbx,
		path: cfg.Path,
	}

	return m, nil
}

func (m *MySQL) String() string {
	return fmt.Sprintf("mysql://%s", m.path)
}

func (m *MySQL) Initialize() error {
	if _, err := m.db.Exec(mysqlSchema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	if err := m.loadItems(); err != nil {
		return fmt.Errorf("load items: %w", err)
	}

	return nil
}

func (m *MySQL) loadItems() error {
	newerThan := time.Now().Add(-1 * MaxLoadAge)

	klog.Infof("loading items from persist table newer than %s ...", newerThan)
	rows, err := m.db.Queryx(`SELECT * FROM persist WHERE saved > ?`, newerThan)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	decoded := map[string]cache.Item{}

	for rows.Next() {
		var mi sqlItem
		err = rows.StructScan(&mi)
		if err != nil {
			return fmt.Errorf("structscan: %w", err)
		}

		var item cache.Item
		gd := gob.NewDecoder(bytes.NewBuffer(mi.Value))
		if err := gd.Decode(&item); err != nil {
			klog.Errorf("decode failed for %s (saved %s): %w", mi.Key, mi.Saved, err)
			continue
		}
		decoded[mi.Key] = item
	}

	klog.Infof("%d items loaded from MySQL", len(decoded))
	m.cache = loadMem(decoded)
	return nil
}

// Set stores a thing
func (m *MySQL) Set(key string, th *Thing) error {
	setMem(m.cache, key, th)

	go func() {
		err := m.persist(key, th)
		if err != nil {
			klog.Errorf("failed to persist %s: %s", key, err)
		}
	}()

	return nil
}

// DeleteOlderThan deletes a thing older than a timestamp
func (m *MySQL) DeleteOlderThan(key string, t time.Time) error {
	deleteOlderMem(m.cache, key, t)
	return nil
}

// GetNewerThan returns a Item older than a timestamp
func (m *MySQL) GetNewerThan(key string, t time.Time) *Thing {
	return newerThanMem(m.cache, key, t)
}

// persist writes an thing to MySQL
func (m *MySQL) persist(key string, th *Thing) error {
	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)
	if err := ge.Encode(th); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	_, err := m.db.Exec(`
		INSERT INTO persist (k, v, saved) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE k=VALUES(k), v=VALUES(v)`, key, b.Bytes(), time.Now())

	return err
}

// Cleanup deletes older cache items
func (m *MySQL) Cleanup() error {
	start := time.Now()
	maxAge := start.Add(-1 * MaxSaveAge)

	res, err := m.db.Exec(`DELETE FROM persist WHERE saved < ?`, maxAge)

	if err != nil {
		return fmt.Errorf("delete exec: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows > 0 {
		klog.Infof("Deleted %d rows of stale data", rows)
	}

	return nil
}
