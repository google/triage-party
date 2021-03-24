// Copyright 2020 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
CREATE TABLE IF NOT EXISTS persist2 (
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
	memcache *cache.Cache
	db       *sqlx.DB
	path     string
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
	m.memcache = createMem()

	if _, err := m.db.Exec(mysqlSchema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

// Set stores a thing
func (m *MySQL) Set(key string, th *Blob) error {
	setMem(m.memcache, key, th)

	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)

	if err := ge.Encode(th); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	_, err := m.db.Exec(`
		INSERT INTO persist2 (k, v, saved) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE k=VALUES(k), v=VALUES(v)`, key, b.Bytes(), time.Now())

	return err
}

// Get returns a Item older than a timestamp
func (m *MySQL) Get(key string, t time.Time) *Blob {
	if b := getMem(m.memcache, key, t); b != nil {
		return b
	}

	rows, err := m.db.Queryx(`SELECT * FROM persist2 WHERE k = ? LIMIT 1`, key)
	if err != nil {
		klog.Errorf("query: %w", err)
		return nil
	}

	for rows.Next() {
		var mi sqlItem
		err = rows.StructScan(&mi)
		if err != nil {
			klog.Errorf("structscan: %w", err)
			return nil
		}

		var bl Blob
		gd := gob.NewDecoder(bytes.NewBuffer(mi.Value))
		if err := gd.Decode(&bl); err != nil {
			klog.Errorf("decode failed for %s (saved %s, bytes: %d): %v", mi.Key, mi.Saved, len(mi.Value), err)
			continue
		}

		if bl.Created.Before(t) {
			klog.Warningf("found %s in db, but it was older than %s", key, t)
			return nil
		}

		setMem(m.memcache, key, &bl)
		return &bl
	}
	return nil
}
