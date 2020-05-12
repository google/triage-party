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
	"time"

	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
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
func (m *Memory) Set(key string, t *Thing) error {
	setMem(m.cache, key, t)
	return nil
}

// DeleteOlderThan deletes a thing older than a timestamp
func (m *Memory) DeleteOlderThan(key string, t time.Time) error {
	deleteOlderMem(m.cache, key, t)
	return nil
}

// GetNewerThan returns a thing older than a timestamp
func (m *Memory) GetNewerThan(key string, t time.Time) *Thing {
	return newerThanMem(m.cache, key, t)
}

func (m *Memory) Save() error {
	klog.Warningf("Save is not implemented by the memory backend")
	return nil
}
