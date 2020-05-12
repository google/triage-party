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
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/store/tikv"
	"k8s.io/klog"
)

const StartKeys = `START_KEYS`

type TikV struct {
	cache *cache.Cache
	cli   *tikv.RawKVClient
	path  string
}

// NewPostgres returns a new Postgres cache
func NewTikV(cfg Config) (*TikV, error) {
	cli, err := tikv.NewRawKVClient([]string{cfg.Path}, config.Security{})
	if err != nil {
		return nil, err
	}
	return &TikV{
		cli:  cli,
		path: cfg.Path,
	}, nil
}

func (t *TikV) String() string {
	return t.path
}

func (t *TikV) Initialize() error {
	klog.Infof("load items")

	if err := t.loadItems(); err != nil {
		return fmt.Errorf("load items: %w", err)
	}

	return nil
}

func (t *TikV) loadItems() error {
	klog.Infof("loading items from persist table ...")
	start, err := t.cli.Get([]byte(StartKeys))
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}
	count, err := strconv.Atoi(string(start))
	if err != nil {
		return fmt.Errorf("convert start: %w", err)
	}
	keys, values, err := t.cli.Scan([]byte(StartKeys), count)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	decoded := map[string]cache.Item{}

	for i := range keys {
		var item cache.Item
		gd := gob.NewDecoder(bytes.NewBuffer(values[i]))
		if err := gd.Decode(&item); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
		decoded[string(keys[i])] = item
	}

	klog.Infof("%d items loaded from Postgres", len(decoded))
	t.cache = loadMem(decoded)
	return nil
}

// Set stores a thing
func (t *TikV) Set(key string, th *Thing) error {
	if key == StartKeys {
		return errors.New(fmt.Sprintf("you can't set key name: %s", StartKeys))
	}
	setMem(t.cache, key, th)
	return nil
}

// DeleteOlderThan deletes a thing older than a timestamp
func (t *TikV) DeleteOlderThan(key string, tt time.Time) error {
	deleteOlderMem(t.cache, key, tt)
	return nil
}

// GetNewerThan returns a Item older than a timestamp
func (t *TikV) GetNewerThan(key string, tt time.Time) *Thing {
	return newerThanMem(t.cache, key, tt)
}

func (t *TikV) Save() error {
	start := time.Now()
	items := t.cache.Items()

	klog.Infof("*** Saving %d items to Tikv", len(items))
	defer func() {
		klog.Infof("*** Tikv.Save took %s", time.Since(start))
	}()
	for k, v := range items {
		b := new(bytes.Buffer)
		ge := gob.NewEncoder(b)
		if err := ge.Encode(v); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
		err := t.cli.Put([]byte(k), b.Bytes())
		if err != nil {
			return err
		}
	}

	return t.cli.Put([]byte(StartKeys), []byte(fmt.Sprintf("%d", len(items))))
}
