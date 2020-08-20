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

package persist

import (
	"time"

	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
)

var (
	memCleanupInterval = 15 * time.Minute
)

func createMem() *cache.Cache {
	return cache.New(MaxLoadAge, memCleanupInterval)
}

func loadMem(items map[string]cache.Item) *cache.Cache {
	for key, v := range items {
		th, ok := v.Object.(*Thing)
		if !ok {
			klog.Warningf("%s is not of type Thing", key)
		} else {
			klog.Infof("found %s (created: %s)", key, th.Created)
		}
	}
	return cache.NewFrom(MaxLoadAge, memCleanupInterval, items)
}

func setMem(c *cache.Cache, key string, th *Thing) {
	if th.Created.IsZero() {
		th.Created = time.Now()
	}

	klog.V(1).Infof("Storing %s within in-memory cache (created: %s)", key, th.Created)
	c.Set(key, th, MaxLoadAge)
}

func newerThanMem(c *cache.Cache, key string, t time.Time) *Thing {
	x, ok := c.Get(key)
	if !ok {
		klog.V(1).Infof("%s is not within in-memory cache!", key)
		return nil
	}

	th, ok := x.(*Thing)
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

func deleteOlderMem(c *cache.Cache, key string, t time.Time) {
	i := newerThanMem(c, key, t)

	// Still good.
	if i != nil && i.Created.After(t) {
		klog.Infof("no need to delete %s", key)
		return
	}

	c.Delete(key)
}
