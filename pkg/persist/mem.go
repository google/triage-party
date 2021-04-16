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
	"time"

	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

var (
	memExpiration      = 7 * 24 * time.Hour
	memCleanupInterval = 12 * time.Hour
)

type Memory struct {
	cache *cache.Cache
}

// NewMemory returns a new Memory cache
func NewMemory(cfg Config) (*Memory, error) {
	return &Memory{}, nil
}

func (m *Memory) String() string {
	return "memory"
}

func (m *Memory) Initialize() error {
	m.cache = createMem()
	return nil
}

// Set stores a thing into memory
func (m *Memory) Set(key string, t *Blob) error {
	setMem(m.cache, key, t)
	return nil
}

// Get returns a thing older than a timestamp
func (m *Memory) Get(key string, t time.Time) *Blob {
	return getMem(m.cache, key, t)
}

func createMem() *cache.Cache {
	return cache.New(memExpiration, memCleanupInterval)
}

func setMem(c *cache.Cache, key string, th *Blob) {
	if th.Created.IsZero() {
		th.Created = time.Now()
	}

	klog.V(1).Infof("Storing %s within in-memory cache (created: %s)", key, th.Created)
	c.Set(key, th, 0)
}

func getMem(c *cache.Cache, key string, t time.Time) *Blob {
	x, ok := c.Get(key)

	if !ok {
		klog.V(1).Infof("%s is not within in-memory cache!", key)
		return nil
	}

	th, ok := x.(*Blob)
	if !ok {
		klog.V(1).Infof("%s is not of type Thing", key)
	}

	if th.Created.After(time.Now()) {
		klog.Errorf("%s claims to be created in the future: %s ???", key, th.Created)
	}

	if th.Created.Before(t) {
		return nil
	}

	return th
}
