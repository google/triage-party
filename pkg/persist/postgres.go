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
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

var pgSchema = `
CREATE TABLE IF NOT EXISTS persist2 (
	id SERIAL PRIMARY KEY,
	saved TIMESTAMP DEFAULT '1970-01-01 00:00:01',
	k VARCHAR UNIQUE,
	v BYTEA
);

CREATE INDEX IF NOT EXISTS saved_idx ON persist2 (saved);
`

type Postgres struct {
	memcache *cache.Cache
	db       *sqlx.DB
	path     string
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
	m.memcache = createMem()

	klog.Infof("schema: %s", pgSchema)
	if _, err := m.db.Exec(pgSchema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

// Set stores a thing
func (m *Postgres) Set(key string, th *Blob) error {
	setMem(m.memcache, key, th)

	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)

	if err := ge.Encode(th); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	_, err := m.db.Exec(`
			INSERT INTO persist2 (k, v, saved) VALUES ($1, $2, $3)
			ON CONFLICT (k)
			DO UPDATE SET v=EXCLUDED.v, saved=EXCLUDED.saved`, key, b.Bytes(), time.Now())

	return err
}

// Get returns a Item older than a timestamp
func (m *Postgres) Get(key string, t time.Time) *Blob {
	start := time.Now()

	if b := getMem(m.memcache, key, t); b != nil {
		return b
	}

	klog.Warningf("%s was not in memory, resorting to SQL cache", key)

	go func() {
		klog.Infof("get(%q) took %s", key, time.Since(start))
	}()

	var mi sqlItem
	err := m.db.Get(&mi, `SELECT * FROM persist2 WHERE k = $1 LIMIT 1`, key)
	if err == sql.ErrNoRows {
		klog.Warningf("%s was not found in SQL table", key)
		return nil
	}

	if err != nil {
		klog.Errorf("query: %w", err)
		return nil
	}

	var bl Blob
	gd := gob.NewDecoder(bytes.NewBuffer(mi.Value))
	if err := gd.Decode(&bl); err != nil {
		klog.Errorf("decode failed for %s (saved %s, bytes: %d): %v", mi.Key, mi.Saved, len(mi.Value), err)
		return nil
	}

	if bl.Created.Before(t) {
		klog.Warningf("found %s in db, but it was older than %s", key, t)
		return nil
	}

	setMem(m.memcache, key, &bl)
	return &bl
}
