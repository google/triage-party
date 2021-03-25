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

// Package persist provides a bootstrap for the in-memory cache
package persist

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

type Disk struct {
	path     string
	subdir   string
	memcache *cache.Cache
	dv       *diskv.Diskv
}

// NewDisk returns a new disk cache
func NewDisk(cfg Config) (*Disk, error) {
	return &Disk{path: cfg.Path, subdir: cfg.Program}, nil
}

func (d *Disk) String() string {
	return d.path
}

func (d *Disk) Initialize() error {
	if d.path == "" {
		root, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("cache dir: %w", err)
		}
		d.path = filepath.Join(root, d.subdir)
	}

	if err := os.MkdirAll(d.path, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	klog.Infof("cache dir is %s", d.path)

	d.memcache = createMem()
	d.dv = diskv.New(diskv.Options{
		BasePath:     d.path,
		CacheSizeMax: 1024 * 1024 * 1024,
	})

	return nil
}

// Set stores a thing into memory
func (d *Disk) Set(key string, bl *Blob) error {
	setMem(d.memcache, key, bl)
	var bs bytes.Buffer
	enc := gob.NewEncoder(&bs)
	err := enc.Encode(bl)
	if err != nil {
		return fmt.Errorf("encode: %v", err)
	}

	return d.dv.Write(key, bs.Bytes())
}

// Get returns a thing older than a timestamp
func (d *Disk) Get(key string, t time.Time) *Blob {
	if b := getMem(d.memcache, key, t); b != nil {
		return b
	}

	if !d.dv.Has(key) {
		klog.Warningf("%s is a complete cache miss", key)
		return nil
	}

	klog.Warningf("%s was not in memory, resorting to disk cache", key)

	var bl Blob
	val, err := d.dv.Read(key)
	if err != nil {
		klog.Errorf("disk read failed for %q: %v", key, err)
		return nil
	}

	enc := gob.NewDecoder(bytes.NewBuffer(val))
	err = enc.Decode(&bl)
	if err != nil {
		klog.Errorf("decode failed for %q: %v", key, err)
		return nil
	}

	if bl.Created.Before(t) {
		klog.Warningf("found %s on disk, but it was older than %s", key, t)
		return nil
	}

	setMem(d.memcache, key, &bl)
	return &bl
}
