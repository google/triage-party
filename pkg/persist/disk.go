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

// Package persist provides a bootstrap for the in-memory cache
package persist

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

	"github.com/google/triage-party/pkg/provider"

	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

type Disk struct {
	path  string
	cache *cache.Cache
}

// NewDisk returns a new disk cache
func NewDisk(cfg Config) (*Disk, error) {
	return &Disk{path: cfg.Path}, nil
}

func (d *Disk) String() string {
	return d.path
}

func (d *Disk) Initialize() error {
	klog.Infof("Initializing with %s ...", d.path)
	if err := d.load(); err != nil {
		klog.Infof("recreating cache due to load error: %v", err)
		d.cache = createMem()
		if err := d.Cleanup(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
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
		return fmt.Errorf("decode failed: %w", err)
	}

	if len(decoded) == 0 {
		return fmt.Errorf("no items on disk")
	}

	klog.Infof("%d items loaded from disk", len(decoded))
	d.cache = loadMem(decoded)
	return nil
}

// Set stores a thing into memory
func (d *Disk) Set(key string, t *provider.Thing) error {
	setMem(d.cache, key, t)
	// Implementation quirk: the disk driver does not persist until Cleanup() is called
	return nil
}

// DeleteOlderThan deletes a thing older than a timestamp
func (d *Disk) DeleteOlderThan(key string, t time.Time) error {
	deleteOlderMem(d.cache, key, t)
	return nil
}

// GetNewerThan returns a thing older than a timestamp
func (d *Disk) GetNewerThan(key string, t time.Time) *provider.Thing {
	return newerThanMem(d.cache, key, t)
}

func (d *Disk) Cleanup() error {
	items := d.cache.Items()
	klog.Infof("*** Saving %d items to disk cache at %s", len(items), d.path)

	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)
	if err := ge.Encode(items); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(d.path), 0o700); err != nil {
		return err
	}

	return ioutil.WriteFile(d.path, b.Bytes(), 0o644)
}

func findCacheRoot() string {
	if _, err := os.Stat("/app/pcache"); err == nil {
		return "/app/pcache"
	}

	if _, err := os.Stat("pcache"); err == nil {
		return "pcache"
	}
	if _, err := os.Stat("../pcache"); err == nil {
		return "../pcache"
	}
	if _, err := os.Stat("../../pcache"); err == nil {
		return "../../pcache"
	}

	cdir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "triage-party")
	}

	return filepath.Join(cdir, "triage-party")
}

func DefaultDiskPath(configPath string, override string) string {
	name := filepath.Base(configPath)
	if override != "" {
		name = name + "_" + strings.Replace(override, "/", "_", -1)
	}

	dir := findCacheRoot()
	path := filepath.Join(dir, name+".pc")
	klog.Infof("default disk path: %s", path)
	return path
}
