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
	"k8s.io/klog"
)

var schema = `
CREATE TABLE IF NOT EXISTS persist (
	id INT AUTO_INCREMENT PRIMARY KEY,
	key VARCHAR(255) NOT NULL,
	ts TIMESTAMP DEFAULT '1970-01-01 00:00:01',
	content MEDIUMBLOB,

	UNIQUE KEY unique_key (key),
	INDEX ts_idx (ts),
);
`

// sqlItem maps to schema
type sqlItem struct {
	ID      int64     `db:"id"`
	Key     string    `db:"key"`
	Created time.Time `db:"created"`
	Saved   time.Time `db:"saved"`
	Value   []byte    `db:"content"`
}

type MySQL struct {
	cache *cache.Cache
	db    *sqlx.DB
}

// NewMySQL returns a new MySQL cache
func NewMySQL(cfg Config) (*MySQL, error) {
	dbx, err := sqlx.Connect("mysql", cfg.Path)
	if err != nil {
		return nil, err
	}

	if _, err := dbx.Exec(schema); err != nil {
		return nil, err
	}

	m := &MySQL{
		db:    dbx,
	}

	if err := m.loadItems(); err != nil {
		return m, fmt.Errorf("load: %w", err)
	}

	return m, nil
}

func (m *MySQL) loadItems() error {
	rows, err := m.db.Queryx(`SELECT * FROM persist`)
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
			return fmt.Errorf("decode: %w", err)
		}
		decoded[mi.Key] = item
	}

	if len(decoded) == 0 {
		return fmt.Errorf("no items loaded from MySQL: %v", decoded)
	}

	klog.Infof("%d items loaded from MySQL", len(decoded))
	m.cache = loadMem(decoded)
	return nil
}

// Set stores a thing
func (m *MySQL) Set(key string, th *Thing) error {
	setMem(m.cache, key, th)
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

func (m *MySQL) Save() error {
	start := time.Now()
	items := m.cache.Items()

	klog.Infof("*** Saving %d items to MySQL", len(items))
	defer func() {
		klog.Infof("*** mysql.Save took %s", time.Since(start))
	}()

	for k, v := range items {
		b := new(bytes.Buffer)
		ge := gob.NewEncoder(b)
		if err := ge.Encode(v); err != nil {
			return fmt.Errorf("encode: %w", err)
		}

		// TODO: figure out how to get th.Created from v
		x, ok := m.cache.Get(k)
		if !ok {
			klog.Errorf("expected %s to be in cache", k)
			continue
		}
		th := x.(*Thing)

		_, err := m.db.Exec(`
			INSERT INTO persist (key, value, ts)
			VALUES (:key, :value, :created, :saved)
			ON DUPLICATE KEY UPDATE
			  value = :value
			  created = :created
			  saved = :saved`,
			sqlItem{Key: k, Value: b.Bytes(), Created: th.Created, Saved: start})

		if err != nil {
			return fmt.Errorf("sql exec: %v (len=%d)", err, len(b.Bytes()))
		}

		return nil
	}

	// Flush older cache items out
	if _, err := m.db.Exec(`DELETE FROM persist WHERE saved < ?`, start); err != nil {
		return err
	}
	return nil
}
