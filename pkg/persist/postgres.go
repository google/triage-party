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
	"github.com/google/triage-party/pkg/provider"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

var pgSchema = `
CREATE TABLE IF NOT EXISTS persist (
	id SERIAL PRIMARY KEY,
	saved TIMESTAMP DEFAULT '1970-01-01 00:00:01',
	k VARCHAR UNIQUE,
	v BYTEA
);

CREATE INDEX IF NOT EXISTS saved_idx ON persist (saved);
`

type Postgres struct {
	cache *cache.Cache
	db    *sqlx.DB
	path  string
}

// NewPostgres returns a new Postgres cache
func NewPostgres(cfg Config) (*Postgres, error) {
	dbx, err := sqlx.Connect("postgres", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	m := &Postgres{
		db:   dbx,
		path: cfg.Path,
	}

	return m, nil
}

func (m *Postgres) String() string {
	return fmt.Sprintf("postgres://%s", m.path)
}

func (m *Postgres) Initialize() error {
	klog.Infof("schema: %s", pgSchema)
	if _, err := m.db.Exec(pgSchema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	if err := m.loadItems(); err != nil {
		return fmt.Errorf("load items: %w", err)
	}

	return nil
}

func (m *Postgres) loadItems() error {
	newerThan := time.Now().Add(-1 * MaxLoadAge)

	klog.Infof("loading items from persist table newer than %s ...", newerThan)
	rows, err := m.db.Queryx(`SELECT * FROM persist WHERE saved > $1`, newerThan)
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
			klog.Errorf("decode failed for %s (saved %s, bytes: %d): %v", mi.Key, mi.Saved, len(mi.Value), err)
			continue
		}
		decoded[mi.Key] = item
	}

	klog.Infof("%d items loaded from Postgres", len(decoded))
	m.cache = loadMem(decoded)
	return nil
}

// Set stores a thing
func (m *Postgres) Set(key string, th *provider.Thing) error {
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
func (m *Postgres) DeleteOlderThan(key string, t time.Time) error {
	deleteOlderMem(m.cache, key, t)
	return nil
}

// GetNewerThan returns a Item older than a timestamp
func (m *Postgres) GetNewerThan(key string, t time.Time) *provider.Thing {
	return newerThanMem(m.cache, key, t)
}

// persist writes an thing to MySQL
func (m *Postgres) persist(key string, th *provider.Thing) error {
	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)

	item := cache.Item{Object: th}
	if err := ge.Encode(item); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	_, err := m.db.Exec(`
			INSERT INTO persist (k, v, saved) VALUES ($1, $2, $3)
			ON CONFLICT (k)
			DO UPDATE SET v=EXCLUDED.v, saved=EXCLUDED.saved`, key, b.Bytes(), time.Now())

	return err
}

// Cleanup deletes older cache items
func (m *Postgres) Cleanup() error {
	start := time.Now()
	maxAge := start.Add(-1 * MaxSaveAge)

	res, err := m.db.Exec(`DELETE FROM persist WHERE saved < $1`, maxAge)

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
