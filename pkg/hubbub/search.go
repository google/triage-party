package hubbub

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hokaccha/go-prettyjson"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/logu"
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

// Search for GitHub issues or PR's
func (h *Engine) SearchAny(ctx context.Context, org string, project string, fs []Filter, newerThan time.Time, hidden bool) ([]*Conversation, time.Time, error) {
	cs, ts, err := h.SearchIssues(ctx, org, project, fs, newerThan, hidden)
	if err != nil {
		return cs, ts, err
	}

	pcs, pts, err := h.SearchPullRequests(ctx, org, project, fs, newerThan, hidden)
	if err != nil {
		return cs, ts, err
	}

	if pts.After(ts) {
		ts = pts
	}

	return append(cs, pcs...), ts, nil
}

// Search for GitHub issues or PR's
func (h *Engine) SearchIssues(ctx context.Context, org string, project string, fs []Filter, newerThan time.Time, hidden bool) ([]*Conversation, time.Time, error) {
	fs = openByDefault(fs)
	klog.Infof("Gathering raw data for %s/%s issues %s - newer than %s", org, project, fs, logu.STime(newerThan))
	var wg sync.WaitGroup

	var open []*github.Issue
	var closed []*github.Issue
	var err error

	age := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		oi, ots, err := h.cachedIssues(ctx, org, project, "open", 0, newerThan)
		if err != nil {
			klog.Errorf("open issues: %v", err)
			return
		}
		if ots.Before(age) {
			age = ots
		}
		open = oi
		klog.V(1).Infof("%s/%s open issue count: %d", org, project, len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !NeedsClosed(fs) {
			return
		}

		ci, cts, err := h.cachedIssues(ctx, org, project, "closed", h.MaxClosedUpdateAge, newerThan)
		if err != nil {
			klog.Errorf("closed issues: %v", err)
		}

		if cts.Before(age) {
			age = cts
		}
		closed = ci

		klog.V(1).Infof("%s/%s closed issue count: %d", org, project, len(closed))
	}()

	wg.Wait()

	var is []*github.Issue
	seen := map[string]bool{}

	for _, i := range append(open, closed...) {
		if len(h.debug) > 0 {
			if h.debug[i.GetNumber()] {
				klog.Errorf("*** Found debug issue #%d:\n%s", i.GetNumber(), formatStruct(*i))
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
	klog.V(1).Infof("%s/%s aggregate issue count: %d, filtering for:\n%s", org, project, len(is), fs)

	// Avoids updating PR references on a quiet repository
	mostRecentUpdate := time.Time{}
	for _, i := range is {
		if i.GetUpdatedAt().After(mostRecentUpdate) {
			mostRecentUpdate = i.GetUpdatedAt()
		}
	}

	for _, i := range is {
		// Inconsistency warning: issues use a list of labels, prs a list of label pointers
		labels := []*github.Label{}
		for _, l := range i.Labels {
			l := l
			labels = append(labels, l)
		}

		if !preFetchMatch(i, labels, fs) {
			klog.V(1).Infof("#%d - %q did not match item filter: %s", i.GetNumber(), i.GetTitle(), fs)
			continue
		}

		klog.V(1).Infof("#%d - %q made it past pre-fetch: %s", i.GetNumber(), i.GetTitle(), fs)

		comments := []*github.IssueComment{}

		fetchComments := false
		if needComments(i, fs) && i.GetComments() > 0 {
			klog.V(1).Infof("#%d - %q: need comments for final filtering", i.GetNumber(), i.GetTitle())
			fetchComments = !newerThan.IsZero()
		}

		comments, _, err = h.cachedIssueComments(ctx, org, project, i.GetNumber(), h.mtime(i), fetchComments)
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		co := h.IssueSummary(i, comments, age)
		co.Labels = labels

		co.Similar = h.FindSimilar(co)
		if len(co.Similar) > 0 {
			co.Tags[tag.Similar] = true
		}

		if !postFetchMatch(co, fs) {
			klog.V(1).Infof("#%d - %q did not match post-fetch filter: %s", i.GetNumber(), i.GetTitle(), fs)
			continue
		}
		klog.V(1).Infof("#%d - %q made it past post-fetch: %s", i.GetNumber(), i.GetTitle(), fs)

		updatedAt := h.mtime(i)
		var timeline []*github.Timeline
		fetchTimeline := false
		if needTimeline(i, fs, false, hidden) {
			fetchTimeline = !newerThan.IsZero()
		}

		timeline, err = h.cachedTimeline(ctx, org, project, i.GetNumber(), updatedAt, fetchTimeline)
		if err != nil {
			klog.Errorf("timeline: %v", err)
		}
		h.addEvents(ctx, co, timeline, fetchTimeline)

		// Some labels are judged by linked PR state. Ensure that they are updated to the same timestamp.
		fetchReviews := false
		if needReviews(i, fs, hidden) && len(co.PullRequestRefs) > 0 {
			fetchReviews = !newerThan.IsZero()
		}
		co.PullRequestRefs = h.updateLinkedPRs(ctx, co, mostRecentUpdate, fetchReviews)

		if !postEventsMatch(co, fs) {
			klog.V(1).Infof("#%d - %q did not match post-events filter: %s", i.GetNumber(), i.GetTitle(), fs)
			continue
		}
		klog.V(1).Infof("#%d - %q made it past post-events: %s", i.GetNumber(), i.GetTitle(), fs)

		filtered = append(filtered, co)
	}

	return filtered, age, nil
}

// NeedsClosed returns whether or not the filters require closed items
func NeedsClosed(fs []Filter) bool {
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
		if f.State != "" && f.State != "open" {
			klog.V(1).Infof("will need closed items due to State=%s", f.State)
			return true
		}
	}
	return false
}

func (h *Engine) SearchPullRequests(ctx context.Context, org string, project string, fs []Filter, newerThan time.Time, hidden bool) ([]*Conversation, time.Time, error) {
	fs = openByDefault(fs)

	klog.Infof("Gathering raw data for %s/%s PR's matching: %s - newer than %s", org, project, fs, logu.STime(newerThan))
	filtered := []*Conversation{}

	var wg sync.WaitGroup

	var open []*github.PullRequest
	var closed []*github.PullRequest
	var err error
	age := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		op, ots, err := h.cachedPRs(ctx, org, project, "open", 0, newerThan)
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
		if !NeedsClosed(fs) {
			return
		}
		cp, cts, err := h.cachedPRs(ctx, org, project, "closed", h.MaxClosedUpdateAge, newerThan)
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

	prs := []*github.PullRequest{}
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
		if !preFetchMatch(pr, pr.Labels, fs) {
			continue
		}

		var timeline []*github.Timeline
		var reviews []*github.PullRequestReview
		var comments []*Comment

		fetchComments := false
		if needComments(pr, fs) {
			fetchComments = !newerThan.IsZero()
		}

		comments, _, err = h.prComments(ctx, org, project, pr.GetNumber(), h.mtime(pr), fetchComments)
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		fetchTimeline := false
		if needTimeline(pr, fs, true, hidden) {
			fetchTimeline = !newerThan.IsZero()
		}

		timeline, err = h.cachedTimeline(ctx, org, project, pr.GetNumber(), h.mtime(pr), fetchTimeline)
		if err != nil {
			klog.Errorf("timeline: %v", err)
		}

		fetchReviews := false
		if needReviews(pr, fs, hidden) {
			fetchReviews = !newerThan.IsZero()
		}
		reviews, _, err = h.cachedReviews(ctx, org, project, pr.GetNumber(), h.mtime(pr), fetchReviews)
		if err != nil {
			klog.Errorf("reviews: %v", err)
			continue
		}

		if h.debug[pr.GetNumber()] {
			klog.Errorf("*** Debug PR timeline #%d:\n%s", pr.GetNumber(), formatStruct(timeline))
		}

		co := h.PRSummary(ctx, pr, comments, timeline, reviews, age, !newerThan.IsZero())
		co.Labels = pr.Labels
		co.Similar = h.FindSimilar(co)
		if len(co.Similar) > 0 {
			co.Tags[tag.Similar] = true
		}

		if !postFetchMatch(co, fs) {
			klog.V(4).Infof("PR #%d did not pass postFetchMatch with filter: %v", pr.GetNumber(), fs)
			continue
		}

		if !postEventsMatch(co, fs) {
			klog.V(1).Infof("#%d - %q did not match post-events filter: %s", pr.GetNumber(), pr.GetTitle(), fs)
			continue
		}

		filtered = append(filtered, co)
	}

	return filtered, age, nil
}

func needComments(i GitHubItem, fs []Filter) bool {
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

	return i.GetState() == "open"
}

func needTimeline(i GitHubItem, fs []Filter, pr bool, hidden bool) bool {
	if i.GetMilestone() != nil {
		return true
	}

	if i.GetState() != "open" {
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

func needReviews(i GitHubItem, fs []Filter, hidden bool) bool {
	if i.GetState() != "open" {
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
