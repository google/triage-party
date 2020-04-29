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

// Package updater handles background refreshes of GitHub data
package updater

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/triage-party/pkg/triage"

	"k8s.io/klog"
)

// Minimum age to flush to avoid bad behavior
const minFlushAge = 1 * time.Second

type PFunc = func() error

type Config struct {
	Party         *triage.Party
	MinRefreshAge time.Duration
	MaxRefreshAge time.Duration
	PersistFunc   PFunc
}

func New(cfg Config) *Updater {
	return &Updater{
		party:         cfg.Party,
		maxRefreshAge: cfg.MaxRefreshAge,
		minRefreshAge: cfg.MinRefreshAge,
		idleDuration:  5 * time.Minute,
		cache:         map[string]*triage.CollectionResult{},
		lastRequest:   sync.Map{},
		loopEvery:     250 * time.Millisecond,
		mutex:         &sync.Mutex{},
		persistFunc:   cfg.PersistFunc,
		startTime:     time.Time{},
	}
}

type Updater struct {
	party         *triage.Party
	maxRefreshAge time.Duration
	minRefreshAge time.Duration
	idleDuration  time.Duration
	cache         map[string]*triage.CollectionResult
	lastRequest   sync.Map
	lastSave      time.Time
	startTime     time.Time
	loopEvery     time.Duration
	mutex         *sync.Mutex
	persistFunc   PFunc
}

// Lookup results for a given metric
func (u *Updater) Lookup(ctx context.Context, id string, blocking bool) *triage.CollectionResult {
	defer u.lastRequest.Store(id, time.Now())
	r := u.cache[id]
	if r == nil {
		if blocking {
			klog.Warningf("%s is not available in the cache, blocking page load!", id)
			if _, err := u.RunSingle(ctx, id, true); err != nil {
				klog.Errorf("unable to run %s: %v", id, err)
			}
		} else {
			klog.Warningf("%s unavailable, but not blocking: happily returning nil", id)
		}
	}
	r = u.cache[id]
	return r
}

func (u *Updater) ForceRefresh(ctx context.Context, id string) *triage.CollectionResult {
	defer u.lastRequest.Store(id, time.Now())

	_, ok := u.lastRequest.Load(id)
	if !ok {
		klog.Warningf("ignoring refresh request, %s has never been requested", id)
		return u.Lookup(ctx, id, true)
	}

	start := time.Now()
	klog.Infof("Forcing refresh for %s", id)
	if err := u.party.FlushSearchCache(id, time.Now().Add(minFlushAge*-1)); err != nil {
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
		klog.Infof("%s is not in cache, needs update", id)
		return true
	}

	resultAge := time.Since(result.Time)
	if resultAge > u.maxRefreshAge {
		klog.Infof("%s is older than max refresh age (%s), should update", id, resultAge)
		return true
	}

	lastRequestAge := time.Since(u.lastRequested(id))
	if resultAge > u.minRefreshAge && lastRequestAge < u.idleDuration {
		klog.Infof("should update %s: %s is older than refresh (%s), but less than idle (%s)", id, resultAge, u.minRefreshAge, u.idleDuration)
		return true
	}
	return false
}

func (u *Updater) lastRequested(id string) time.Time {
	x, ok := u.lastRequest.Load(id)
	if !ok {
		return u.startTime
	}

	lr, ok := x.(time.Time)
	if !ok {
		return u.startTime
	}

	return lr
}

func (u *Updater) update(ctx context.Context, s triage.Collection) error {
	if u.lastSave.IsZero() {
		klog.Infof("have not yet saved content - will accept stale results")
		u.party.AcceptStaleResults(true)
	} else {
		u.party.AcceptStaleResults(false)
	}

	r, err := u.party.ExecuteCollection(ctx, s, u.lastRequested(s.ID))
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

	s, err := u.party.LookupCollection(id)
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
	sts, err := u.party.ListCollections()
	if err != nil {
		return err
	}

	var failed []string
	for _, s := range sts {
		runUpdated, err := u.RunSingle(ctx, s.ID, force)
		if err != nil {
			klog.Errorf("%s failed to update: %v", s.ID, err)
			failed = append(failed, s.ID)
		}
		if runUpdated {
			updated = true
		}
	}

	if updated && time.Since(u.lastSave) > u.maxRefreshAge {
		if err := u.persistFunc(); err != nil {
			klog.Errorf("persist failed: %v", err)
		} else {
			u.lastSave = time.Now()
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("collections failed: %v", failed)
	}

	return nil
}

// Update loop
func (u *Updater) Loop(ctx context.Context) error {
	klog.Infof("Looping: data will be updated between %s and %s", u.minRefreshAge, u.maxRefreshAge)

	// Quickly establish a baseline with stale data
	if err := u.RunOnce(ctx, false); err != nil {
		return err
	}

	u.startTime = time.Now()

	// Run once with fresh data
	if err := u.RunOnce(ctx, true); err != nil {
		return err
	}

	// Loop if everything goes to plan
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
