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

	"github.com/google/triage-party/pkg/logu"
	"github.com/patrickmn/go-cache"
	"k8s.io/klog"
)

var (
	memExpireInterval  = 65 * 24 * time.Hour
	memCleanupInterval = 15 * time.Minute
)

func createMem() *cache.Cache {
	return cache.New(memExpireInterval, memCleanupInterval)
}

func loadMem(items map[string]cache.Item) *cache.Cache {
	return cache.NewFrom(memExpireInterval, memCleanupInterval, items)
}

func setMem(c *cache.Cache, key string, th *Thing) {
	if th.Created.IsZero() {
		th.Created = time.Now()
	}

	klog.V(1).Infof("Storing %s within in-memory cache", key)
	c.Set(key, th, memExpireInterval)
}

func newerThanMem(c *cache.Cache, key string, t time.Time) *Thing {
	x, ok := c.Get(key)
	if !ok {
		klog.Infof("%s is not within in-memory cache!", key)
		return nil
	}

	th := x.(*Thing)

	if th.Created.Before(t) {
		klog.V(2).Infof("%s in cache, but %s is older than %s", key, logu.STime(th.Created), logu.STime(t))
		return nil
	}

	return th
}

func deleteOlderMem(c *cache.Cache, key string, t time.Time) {
	i := newerThanMem(c, key, t)

	// Still good.
	if i != nil && i.Created.After(t) {
		klog.Infof("no need to delete %s", key)
		return
	}

	c.Delete(key)
}
