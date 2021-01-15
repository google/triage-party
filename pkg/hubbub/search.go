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
	"strings"
	"sync"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"

	"github.com/davecgh/go-spew/spew"
	"github.com/hokaccha/go-prettyjson"

	"github.com/google/triage-party/pkg/logu"
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

// Search for GitHub issues or PR's
func (h *Engine) SearchAny(ctx context.Context, sp provider.SearchParams) ([]*Conversation, time.Time, error) {
	cs, ts, err := h.SearchIssues(ctx, sp)
	if err != nil {
		return cs, ts, err
	}

	pcs, pts, err := h.SearchPullRequests(ctx, sp)
	if err != nil {
		return cs, ts, err
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
	var err error

	age := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sp.State = constants.OpenState
		if sp.Repo.Host == constants.GitlabProviderHost {
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
	mostRecentUpdate := time.Time{}
	for _, i := range is {
		if i.GetUpdatedAt().After(mostRecentUpdate) {
			mostRecentUpdate = i.GetUpdatedAt()
		}
	}

	for _, i := range is {
		// Inconsistency warning: issues use a list of labels, prs a list of label pointers
		labels := []*provider.Label{}
		for _, l := range i.Labels {
			l := l
			labels = append(labels, l)
		}

		if !preFetchMatch(i, labels, sp.Filters) {
			klog.V(1).Infof("#%d - %q did not match item filter: %s", i.GetNumber(), i.GetTitle(), sp.Filters)
			continue
		}

		klog.V(1).Infof("#%d - %q made it past pre-fetch: %s", i.GetNumber(), i.GetTitle(), sp.Filters)

		comments := []*provider.IssueComment{}

		fetchComments := false
		if needComments(i, sp.Filters) && i.GetComments() > 0 {
			klog.V(1).Infof("#%d - %q: need comments for final filtering", i.GetNumber(), i.GetTitle())
			fetchComments = !sp.NewerThan.IsZero()
		}

		sp.IssueNumber = i.GetNumber()
		sp.NewerThan = h.mtime(i)
		sp.Fetch = fetchComments

		comments, _, err = h.cachedIssueComments(ctx, sp)
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		co := h.IssueSummary(i, comments, age)
		co.Labels = labels

		co.Similar = h.FindSimilar(co)
		if len(co.Similar) > 0 {
			co.Tags[tag.Similar] = true
		}

		if !postFetchMatch(co, sp.Filters) {
			klog.V(1).Infof("#%d - %q did not match post-fetch filter: %s", i.GetNumber(), i.GetTitle(), sp.Filters)
			continue
		}
		klog.V(1).Infof("#%d - %q made it past post-fetch: %s", i.GetNumber(), i.GetTitle(), sp.Filters)

		updatedAt := h.mtime(i)
		var timeline []*provider.Timeline
		fetchTimeline := false
		if needTimeline(i, sp.Filters, false, sp.Hidden) {
			fetchTimeline = !sp.NewerThan.IsZero()
		}

		sp.IssueNumber = i.GetNumber()
		sp.Fetch = fetchTimeline
		sp.UpdateAt = updatedAt

		timeline, err = h.cachedTimeline(ctx, sp)
		if err != nil {
			klog.Errorf("timeline: %v", err)
		}

		h.addEvents(ctx, sp, co, timeline)

		// Some labels are judged by linked PR state. Ensure that they are updated to the same timestamp.
		fetchReviews := false
		if needReviews(i, sp.Filters, sp.Hidden) && len(co.PullRequestRefs) > 0 {
			fetchReviews = !sp.NewerThan.IsZero()
		}
		sp.NewerThan = mostRecentUpdate
		sp.Fetch = fetchReviews
		co.PullRequestRefs = h.updateLinkedPRs(ctx, sp, co)

		if !postEventsMatch(co, sp.Filters) {
			klog.V(1).Infof("#%d - %q did not match post-events filter: %s", i.GetNumber(), i.GetTitle(), sp.Filters)
			continue
		}
		klog.V(1).Infof("#%d - %q made it past post-events: %s", i.GetNumber(), i.GetTitle(), sp.Filters)

		filtered = append(filtered, co)
	}

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
	var err error
	age := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sp.State = constants.OpenState
		if sp.Repo.Host == constants.GitlabProviderHost {
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
		if !preFetchMatch(pr, pr.Labels, sp.Filters) {
			continue
		}

		var timeline []*provider.Timeline
		var reviews []*provider.PullRequestReview
		var comments []*provider.Comment

		fetchComments := false
		if needComments(pr, sp.Filters) {
			fetchComments = !sp.NewerThan.IsZero()
		}

		sp.IssueNumber = pr.GetNumber()
		sp.NewerThan = h.mtime(pr)
		sp.Fetch = fetchComments

		comments, _, err = h.prComments(ctx, sp)
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		fetchTimeline := false
		if needTimeline(pr, sp.Filters, true, sp.Hidden) {
			fetchTimeline = !sp.NewerThan.IsZero()
		}

		sp.IssueNumber = pr.GetNumber()
		sp.NewerThan = h.mtime(pr)
		sp.Fetch = fetchTimeline

		timeline, err = h.cachedTimeline(ctx, sp)
		if err != nil {
			klog.Errorf("timeline: %v", err)
		}

		fetchReviews := false
		if needReviews(pr, sp.Filters, sp.Hidden) {
			fetchReviews = !sp.NewerThan.IsZero()
		}

		sp.IssueNumber = pr.GetNumber()
		sp.NewerThan = h.mtime(pr)
		sp.Fetch = fetchReviews

		reviews, _, err = h.cachedReviews(ctx, sp)
		if err != nil {
			klog.Errorf("reviews: %v", err)
			continue
		}

		if h.debug[pr.GetNumber()] {
			klog.Errorf("*** Debug PR timeline #%d:\n%s", pr.GetNumber(), formatStruct(timeline))
		}

		sp.Fetch = !sp.NewerThan.IsZero()
		sp.Age = age

		co := h.PRSummary(ctx, sp, pr, comments, timeline, reviews)
		co.Labels = pr.Labels
		co.Similar = h.FindSimilar(co)
		if len(co.Similar) > 0 {
			co.Tags[tag.Similar] = true
		}

		if !postFetchMatch(co, sp.Filters) {
			klog.V(4).Infof("PR #%d did not pass postFetchMatch with filter: %v", pr.GetNumber(), sp.Filters)
			continue
		}

		if !postEventsMatch(co, sp.Filters) {
			klog.V(1).Infof("#%d - %q did not match post-events filter: %s", pr.GetNumber(), pr.GetTitle(), sp.Filters)
			continue
		}

		filtered = append(filtered, co)
	}

	return filtered, age, nil
}

func needComments(i provider.IItem, fs []provider.Filter) bool {
	for _, f := range fs {
		if f.TagRegex() != nil {
			if ok, t := matchTag(tag.Tags, f.TagRegex(), f.TagNegate()); ok {
				if t.NeedsComments {
					klog.Infof("#%d - need comments due to tag %s (negate=%v)", i.GetNumber(), f.TagRegex(), f.TagNegate())
					return true
				}
			}
		}

		if f.ClosedCommenters != "" || f.ClosedComments != "" {
			klog.Infof("#%d - need comments due to closed comments", i.GetNumber())
			return true
		}

		if f.Responded != "" || f.Commenters != "" {
			klog.Infof("#%d - need comments due to responded/commenters filter", i.GetNumber())
			return true
		}
	}

	return (i.GetState() == constants.OpenState) || (i.GetState() == constants.OpenedState)
}

func needTimeline(i provider.IItem, fs []provider.Filter, pr bool, hidden bool) bool {
	if i.GetMilestone() != nil {
		return true
	}

	if (i.GetState() != constants.OpenState) && (i.GetState() != constants.OpenedState) {
		return false
	}

	if i.GetUpdatedAt() == i.GetCreatedAt() {
		return false
	}

	if pr {
		return true
	}

	for _, f := range fs {
		if f.TagRegex() != nil {
			if ok, t := matchTag(tag.Tags, f.TagRegex(), f.TagNegate()); ok {
				if t.NeedsTimeline {
					return true
				}
			}
		}
		if f.Prioritized != "" {
			return true
		}
	}

	return !hidden
}

func needReviews(i provider.IItem, fs []provider.Filter, hidden bool) bool {
	if (i.GetState() != constants.OpenState) && (i.GetState() != constants.OpenedState) {
		return false
	}

	if i.GetUpdatedAt() == i.GetCreatedAt() {
		return false
	}

	if hidden {
		return false
	}

	for _, f := range fs {
		if f.TagRegex() != nil {
			if ok, t := matchTag(tag.Tags, f.TagRegex(), f.TagNegate()); ok {
				if t.NeedsReviews {
					klog.V(1).Infof("#%d - need reviews due to tag %s (negate=%v)", i.GetNumber(), f.TagRegex(), f.TagNegate())
					return true
				}
			}
		}
	}

	return true
}

func formatStruct(x interface{}) string {
	s, err := prettyjson.Marshal(x)
	if err == nil {
		return string(s)
	}
	y := strings.Replace(spew.Sdump(x), "\n", "\n|", -1)
	y = strings.Replace(y, ", ", ",\n - ", -1)
	y = strings.Replace(y, "}, ", "},\n", -1)
	return strings.Replace(y, "},\n - ", "},\n", -1)
}
