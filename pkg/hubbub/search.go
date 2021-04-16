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

package hubbub

import (
	"context"
	"sync"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"

	"github.com/google/triage-party/pkg/logu"
	"k8s.io/klog/v2"
)

// Search for GitHub issues or PR's
func (h *Engine) SearchAny(ctx context.Context, sp provider.SearchParams) ([]*Conversation, time.Time, error) {
	var wg sync.WaitGroup
	var cs []*Conversation
	var ts time.Time
	var err error

	wg.Add(1)
	go func() {
		cs, ts, err = h.SearchIssues(ctx, sp)
		wg.Done()
	}()

	var pcs []*Conversation
	var pts time.Time
	var perr error

	wg.Add(1)
	go func() {
		pcs, pts, perr = h.SearchPullRequests(ctx, sp)
		wg.Done()
	}()

	wg.Wait()

	if err != nil {
		return cs, ts, err
	}

	if perr != nil {
		return pcs, pts, perr
	}

	if pts.After(ts) {
		ts = pts
	}

	return append(cs, pcs...), ts, nil
}

// Search for GitHub issues or PR's
func (h *Engine) SearchIssues(ctx context.Context, sp provider.SearchParams) ([]*Conversation, time.Time, error) {
	sp.Filters = openByDefault(sp)
	klog.V(1).Infof(
		"Gathering raw data for %s/%s issues %s - newer than %s",
		sp.Repo.Organization,
		sp.Repo.Project,
		sp.Filters,
		logu.STime(sp.NewerThan),
	)
	var wg sync.WaitGroup

	var open []*provider.Issue
	var closed []*provider.Issue

	start := time.Now()
	age := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sp.State = constants.OpenState
		if sp.Repo.Host == constants.GitLabProviderHost {
			sp.State = constants.OpenedState
		}

		oi, ots, err := h.cachedIssues(ctx, sp)
		if err != nil {
			klog.Errorf("open issues: %v", err)
			return
		}
		if ots.Before(age) {
			age = ots
		}
		open = oi
		klog.V(1).Infof("%s/%s open issue count: %d", sp.Repo.Organization, sp.Repo.Project, len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !NeedsClosed(sp.Filters) {
			return
		}

		sp.State = constants.ClosedState
		sp.UpdateAge = h.MaxClosedUpdateAge

		ci, cts, err := h.cachedIssues(ctx, sp)
		if err != nil {
			klog.Errorf("closed issues: %v", err)
		}

		if cts.Before(age) {
			age = cts
		}
		closed = ci

		klog.V(1).Infof("%s/%s closed issue count: %d", sp.Repo.Organization, sp.Repo.Project, len(closed))
	}()

	wg.Wait()

	var is []*provider.Issue
	seen := map[string]bool{}

	for _, i := range append(open, closed...) {
		if len(h.debug) > 0 {
			if h.debug[i.GetNumber()] {
				klog.Errorf("*** Found debug issue #%d:\n%s", i.GetNumber(), formatStruct(i))
			} else {
				continue
			}
		}

		if seen[i.GetURL()] {
			klog.Errorf("unusual: I already saw #%d", i.GetURL())
			continue
		}
		seen[i.GetURL()] = true
		is = append(is, i)
	}

	var filtered []*Conversation
	klog.V(1).Infof("%s/%s aggregate issue count: %d, filtering for:\n%s", sp.Repo.Organization, sp.Repo.Project, len(is), sp.Filters)

	// Avoids updating PR references on a quiet repository
	latestIssueUpdate := time.Time{}
	for _, i := range is {
		if i.GetUpdatedAt().After(latestIssueUpdate) {
			latestIssueUpdate = i.GetUpdatedAt()
		}
	}

	for _, i := range is {
		if co := h.analyzeIssue(ctx, i, sp, age, latestIssueUpdate); co != nil {
			filtered = append(filtered, co)
		}
	}

	klog.Infof("issue search took %s, returning %d items: %+v", time.Since(start), len(filtered), sp)
	return filtered, age, nil
}

// NeedsClosed returns whether or not the filters require closed items
func NeedsClosed(fs []provider.Filter) bool {
	// First-pass filter: do any filters require closed data?
	for _, f := range fs {
		if f.ClosedCommenters != "" {
			klog.V(1).Infof("will need closed items due to ClosedCommenters=%s", f.ClosedCommenters)
			return true
		}
		if f.ClosedComments != "" {
			klog.V(1).Infof("will need closed items due to ClosedComments=%s", f.ClosedComments)
			return true
		}
		if f.State != "" && ((f.State != constants.OpenState) && (f.State != constants.OpenedState)) {
			klog.V(1).Infof("will need closed items due to State=%s", f.State)
			return true
		}
	}
	return false
}

func (h *Engine) SearchPullRequests(ctx context.Context, sp provider.SearchParams) ([]*Conversation, time.Time, error) {
	sp.Filters = openByDefault(sp)

	klog.V(1).Infof("Gathering raw data for %s/%s PR's matching: %s - newer than %s",
		sp.Repo.Organization, sp.Repo.Project, sp.Filters, logu.STime(sp.NewerThan))
	filtered := []*Conversation{}

	var wg sync.WaitGroup

	var open []*provider.PullRequest
	var closed []*provider.PullRequest
	age := time.Now()
	start := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sp.State = constants.OpenState
		if sp.Repo.Host == constants.GitLabProviderHost {
			sp.State = constants.OpenedState
		}
		sp.UpdateAge = 0

		op, ots, err := h.cachedPRs(ctx, sp)
		if err != nil {
			klog.Errorf("open prs: %v", err)
			return
		}
		if ots.Before(age) {
			klog.Infof("setting age to %s (open PR count)", ots)
			age = ots
		}
		open = op
		klog.V(1).Infof("open PR count: %d", len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !NeedsClosed(sp.Filters) {
			return
		}

		sp.UpdateAge = h.MaxClosedUpdateAge
		sp.State = constants.ClosedState

		cp, cts, err := h.cachedPRs(ctx, sp)
		if err != nil {
			klog.Errorf("closed prs: %v", err)
			return
		}

		if cts.Before(age) {
			klog.Infof("setting age to %s (open PR count)", cts)
			age = cts
		}

		closed = cp

		klog.V(1).Infof("closed PR count: %d", len(closed))
	}()

	wg.Wait()

	prs := []*provider.PullRequest{}
	for _, pr := range append(open, closed...) {
		if len(h.debug) > 0 {
			if h.debug[pr.GetNumber()] {
				klog.Errorf("*** Found debug PR #%d:\n%s", pr.GetNumber(), formatStruct(*pr))
			} else {
				klog.V(2).Infof("Ignoring #%s - does not match debug filter: %v", pr.GetHTMLURL(), h.debug)
				continue
			}
		}
		prs = append(prs, pr)
	}

	for _, pr := range prs {
		if co := h.analyzePR(ctx, pr, sp, age); co != nil {
			filtered = append(filtered, co)
		}
	}

	klog.Infof("PR search took %s, returning %d items: %+v", time.Since(start), len(filtered), sp)
	return filtered, age, nil
}
