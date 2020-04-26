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

// updater package handles background updates of GitHub data
package updater

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/triage-party/pkg/hubbub"

	"github.com/golang/glog"
	"github.com/google/go-github/v24/github"
	"k8s.io/klog"
)

// Minimum age to flush to avoid bad behavior
const minFlushAge = 1 * time.Second

type PFunc = func() error

type Config struct {
	HubBub        *hubbub.HubBub
	Client        *github.Client
	MinRefreshAge time.Duration
	MaxRefreshAge time.Duration
	PersistFunc   PFunc
}

func New(cfg Config) *Updater {
	return &Updater{
		hubbub:        cfg.HubBub,
		client:        cfg.Client,
		maxRefreshAge: cfg.MaxRefreshAge,
		minRefreshAge: cfg.MinRefreshAge,
		cache:         map[string]*hubbub.CollectionResult{},
		lastRequest:   sync.Map{},
		loopEvery:     250 * time.Millisecond,
		mutex:         &sync.Mutex{},
		persistFunc:   cfg.PersistFunc,
	}
}

type Updater struct {
	hubbub        *hubbub.HubBub
	client        *github.Client
	maxRefreshAge time.Duration
	minRefreshAge time.Duration
	cache         map[string]*hubbub.CollectionResult
	lastRequest   sync.Map
	lastSave      time.Time
	loopEvery     time.Duration
	mutex         *sync.Mutex
	persistFunc   PFunc
}

// Lookup results for a given metric
func (u *Updater) Lookup(ctx context.Context, id string, blocking bool) *hubbub.CollectionResult {
	defer u.lastRequest.Store(id, time.Now())
	r := u.cache[id]
	if r == nil {
		if blocking {
			glog.Warningf("%s unavailable, blocking page load!", id)
			if _, err := u.RunSingle(ctx, id, true); err != nil {
				glog.Errorf("unable to run %s: %v", id, err)
			}
		} else {
			glog.Warningf("%s unavailable, but not blocking: happily returning nil", id)
		}
	}
	r = u.cache[id]
	return r
}

func (u *Updater) ForceRefresh(ctx context.Context, id string) *hubbub.CollectionResult {
	defer u.lastRequest.Store(id, time.Now())

	_, ok := u.lastRequest.Load(id)
	if !ok {
		klog.Warningf("ignoring refresh request, %s has never been requested", id)
		return u.Lookup(ctx, id, true)
	}

	start := time.Now()
	klog.Infof("Forcing refresh for %s", id)
	if err := u.hubbub.FlushSearchCache(id, minFlushAge); err != nil {
		klog.Errorf("unable to flush cache: %v", err)
	}

	if _, err := u.RunSingle(ctx, id, true); err != nil {
		klog.Errorf("update failed: %v", err)
	}
	klog.Infof("refresh complete for %s after %s", id, time.Since(start))
	return u.cache[id]
}

func (u *Updater) shouldUpdate(id string) bool {
	result, ok := u.cache[id]
	if !ok {
		return true
	}
	age := time.Since(result.Time)
	if age > u.maxRefreshAge {
		klog.Infof("%s is too old (%s), refreshing", id, age)
		return true
	}

	lastReq, ok := u.lastRequest.Load(id)
	if !ok {
		return false
	}
	lr, ok := lastReq.(time.Time)
	if !ok {
		return false
	}

	if lr.After(result.Time) && age > u.minRefreshAge {
		klog.Infof("%s not updated since last request (%s), refreshing", id, age)
		return true
	}
	return false
}

func (u *Updater) update(ctx context.Context, s hubbub.Collection) error {
	r, err := u.hubbub.ExecuteCollection(ctx, u.client, s)
	if err != nil {
		return err
	}
	u.cache[s.ID] = r
	return nil
}

// Run a single collection, optionally forcing an update
func (u *Updater) RunSingle(ctx context.Context, id string, force bool) (bool, error) {
	updated := false
	klog.V(3).Infof("RunSingle: %s (locking mutex)", id)
	u.mutex.Lock()
	defer u.mutex.Unlock()

	s, err := u.hubbub.LookupCollection(id)
	if err != nil {
		return updated, err
	}

	if force || u.shouldUpdate(s.ID) {
		klog.Infof("must update: %s", s.ID)
		err := u.update(ctx, s)
		if err != nil {
			return updated, err
		}
		updated = true
	}
	return updated, nil
}

// Run once, optionally forcing an update
func (u *Updater) RunOnce(ctx context.Context, force bool) error {
	updated := false
	klog.V(3).Infof("RunOnce: force=%v", force)
	sts, err := u.hubbub.ListCollections()
	if err != nil {
		return err
	}

	var failed []string
	for _, s := range sts {
		u, err := u.RunSingle(ctx, s.ID, force)
		if err != nil {
			klog.Errorf("%s failed to update: %v", s.ID, err)
			failed = append(failed, s.ID)
		}
		if u {
			updated = u
		}
		continue
	}

	if updated && time.Since(u.lastSave) > u.maxRefreshAge {
		u.persistFunc()
		u.lastSave = time.Now()
	}

	if len(failed) > 0 {
		return fmt.Errorf("collections failed: %v", failed)
	}
	return nil
}

// Update loop
func (u *Updater) Loop(ctx context.Context) error {
	ticker := time.NewTicker(u.loopEvery)
	defer ticker.Stop()
	for range ticker.C {
		err := u.RunOnce(ctx, false)
		if err != nil {
			klog.Errorf("err: %v", err)
		}
	}
	return nil
}
