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

// Package initcache provides a bootstrap for the in-memory cache

package initcache

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/logu"
	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

const (
	DiskExpireInterval  = 65 * 24 * time.Hour
	DiskCleanupInterval = 15 * time.Minute
)

type Disk struct {
	path  string
	cache *cache.Cache
}

// NewDisk returns a new disk cache
func NewDisk(cfg Config) *Disk {
	gob.Register(&Hoard{})
	return &Disk{path: cfg.Path}
}

// Initialize creates or loads the disk cache
func (d *Disk) Initialize() error {
	klog.Infof("Initializing with %s ...", d.path)
	if err := d.load(); err != nil {
		klog.Infof("recreating cache due to load error: %v", err)
		return d.create()
	}
	return nil
}

func (d *Disk) load() error {
	f, err := os.Open(d.path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	decoded := map[string]cache.Item{}

	gd := gob.NewDecoder(bufio.NewReader(f))

	err = gd.Decode(&decoded)
	if err != nil && err != io.EOF {
		klog.Errorf("Decode failed: %v", err)
		return d.create()
	}

	if len(decoded) == 0 {
		return fmt.Errorf("no items loaded from disk: %v", decoded)
	}

	klog.Infof("%d items loaded from disk", len(decoded))
	d.cache = cache.NewFrom(DiskExpireInterval, DiskCleanupInterval, decoded)
	return nil
}

// Set stores a hoard onto disk
func (d *Disk) Set(key string, h *Hoard) error {
	if h.Creation.IsZero() {
		h.Creation = time.Now()
	}

	klog.V(1).Infof("Storing %s within in-memory cache", key)
	d.cache.Set(key, h, DiskExpireInterval)
	return nil
}

// DeleteOlderThan deletes a hoard older than a timestamp
func (d *Disk) DeleteOlderThan(key string, t time.Time) error {
	d.cache.Delete(key)
	return nil
}

// GetNewerThan returns a hoard older than a timestamp
func (d *Disk) GetNewerThan(key string, t time.Time) *Hoard {
	x, ok := d.cache.Get(key)
	if !ok {
		klog.Infof("%s is not within in-memory cache!", key)
		return nil
	}

	h := x.(*Hoard)

	if h.Creation.Before(t) {
		klog.V(2).Infof("%s in cache, but %s is older than %s", key, logu.STime(h.Creation), logu.STime(t))
		return nil
	}

	return h
}

func (d *Disk) create() error {
	klog.Infof("Creating in-memory cache, expire interval: %s", DiskExpireInterval)

	d.cache = cache.New(DiskExpireInterval, DiskCleanupInterval)
	if err := d.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (d *Disk) Save() error {
	start := time.Now()
	items := d.cache.Items()

	klog.Infof("*** Saving %d items to disk cache at %s", len(items), d.path)
	defer func() {
		klog.Infof("*** disk.Save took %s", time.Since(start))
	}()

	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)
	if err := ge.Encode(items); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return ioutil.WriteFile(d.path, b.Bytes(), 0644)
}

func DefaultDiskPath(configPath string, override string) string {
	name := strings.Replace(filepath.Base(configPath), filepath.Ext(configPath), "", -1)

	if override != "" {
		name = name + "_" + strings.Replace(override, "/", "_", -1)
	}

	// os.UserCacheDir() is technically better, but difficult to calculate in Dockerfile
	home, err := os.UserHomeDir()
	if err != nil {
		klog.Exitf("unable to get home directory: %v", err)
	}

	return filepath.Join(home, ".tpcache", name)
}
