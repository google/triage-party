package hubbub

import (
	"context"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/hokaccha/go-prettyjson"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

// Search for GitHub issues or PR's
func (h *Engine) SearchAny(ctx context.Context, org string, project string, fs []Filter) ([]*Conversation, error) {
	cs, err := h.SearchIssues(ctx, org, project, fs)
	if err != nil {
		return cs, err
	}

	pcs, err := h.SearchPullRequests(ctx, org, project, fs)
	if err != nil {
		return cs, err
	}

	return append(cs, pcs...), nil
}

// Search for GitHub issues or PR's
func (h *Engine) SearchIssues(ctx context.Context, org string, project string, fs []Filter) ([]*Conversation, error) {
	fs = openByDefault(fs)
	klog.Infof("Gathering raw data for %s/%s search:\n%s", org, project, toYAML(fs))
	var wg sync.WaitGroup

	var members map[string]bool
	var open []*github.Issue
	var closed []*github.Issue
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()
		members, err = h.cachedOrgMembers(ctx, org)
		if err != nil {
			klog.Errorf("members: %v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		open, err = h.cachedIssues(ctx, org, project, "open", 0)
		if err != nil {
			klog.Errorf("open issues: %v", err)
			return
		}
		klog.V(1).Infof("%s/%s open issue count: %d", org, project, len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		closed, err = h.cachedIssues(ctx, org, project, "closed", closedIssueDays)
		if err != nil {
			klog.Errorf("closed issues: %v", err)
		}
		klog.V(1).Infof("%s/%s closed issue count: %d", org, project, len(closed))
	}()

	wg.Wait()

	var is []*github.Issue
	seen := map[string]bool{}

	for _, i := range append(open, closed...) {
		if h.debugNumber != 0 {
			if i.GetNumber() == h.debugNumber {
				klog.Errorf("*** Found debug issue #%d:\n%s", i.GetNumber(), formatStruct(*i))

			} else {
				continue
			}
		}

		if seen[i.GetURL()] {
			klog.Errorf("unusual: I already saw #%d", i.GetNumber())
			continue
		}
		seen[i.GetURL()] = true
		is = append(is, i)
	}

	var filtered []*Conversation
	klog.V(1).Infof("%s/%s aggregate issue count: %d, filtering for:\n%s", org, project, len(is), toYAML(fs))

	for _, i := range is {
		// Inconsistency warning: issues use a list of labels, prs a list of label pointers
		labels := []*github.Label{}
		for _, l := range i.Labels {
			l := l
			labels = append(labels, l)
		}

		if !preFetchMatch(i, labels, fs) {
			klog.V(1).Infof("#%d - %q did not match item filter: %s", i.GetNumber(), i.GetTitle(), toYAML(fs))
			continue
		}

		comments := []*github.IssueComment{}
		if i.GetComments() > 0 {
			klog.V(1).Infof("#%d - %q: need comments for final filtering", i.GetNumber(), i.GetTitle())
			comments, err = h.cachedIssueComments(ctx, org, project, i.GetNumber(), i.GetUpdatedAt())
			if err != nil {
				klog.Errorf("comments: %v", err)
			}
		}

		co := h.IssueSummary(i, comments, members[i.User.GetLogin()])
		co.Labels = labels
		h.seen[co.URL] = co

		if !postFetchMatch(co, fs) {
			klog.V(1).Infof("#%d - %q did not match conversation filter: %s", i.GetNumber(), i.GetTitle(), toYAML(fs))
			continue
		}

		filtered = append(filtered, co)
	}

	// TODO: Make this only happen when caches are missed
	if err := h.updateSimilarConversations(filtered); err != nil {
		klog.Errorf("update similar: %v", err)
	}

	klog.Infof("%d of %d issues within %s/%s matched filters:\n%s", len(filtered), len(is), org, project, toYAML(fs))
	return filtered, nil
}

func (h *Engine) SearchPullRequests(ctx context.Context, org string, project string, fs []Filter) ([]*Conversation, error) {
	fs = openByDefault(fs)

	klog.Infof("Searching %s/%s for PR's matching: %s", org, project, toYAML(fs))
	filtered := []*Conversation{}

	var wg sync.WaitGroup

	var open []*github.PullRequest
	var closed []*github.PullRequest
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()
		open, err = h.cachedPRs(ctx, org, project, "open", 0)
		if err != nil {
			klog.Errorf("open prs: %v", err)
			return
		}
		klog.V(1).Infof("open PR count: %d", len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		closed, err = h.cachedPRs(ctx, org, project, "closed", closedPRDays)
		if err != nil {
			klog.Errorf("closed prs: %v", err)
			return
		}
		klog.V(1).Infof("closed PR count: %d", len(closed))
	}()

	wg.Wait()

	prs := []*github.PullRequest{}
	for _, pr := range append(open, closed...) {
		if h.debugNumber != 0 {
			if pr.GetNumber() == h.debugNumber {
				klog.Errorf("*** Found debug PR #%d:\n%s", pr.GetNumber(), formatStruct(*pr))
			} else {
				continue
			}
		}
		prs = append(prs, pr)
	}

	for _, pr := range prs {
		klog.V(3).Infof("Found PR #%d with labels: %+v", pr.GetNumber(), pr.Labels)
		if !preFetchMatch(pr, pr.Labels, fs) {
			klog.V(4).Infof("PR #%d did not pass preFetchMatch :(", pr.GetNumber())
			continue
		}

		comments := []*github.PullRequestComment{}
		// pr.GetComments() always returns 0 :(
		if pr.GetState() == "open" && pr.GetUpdatedAt().After(pr.GetCreatedAt()) {
			comments, err = h.cachedPRComments(ctx, org, project, pr.GetNumber(), pr.GetUpdatedAt())
			if err != nil {
				klog.Errorf("comments: %v", err)
			}
			if pr.GetNumber() == h.debugNumber {
				klog.Errorf("debug comments: %s", formatStruct(comments))
			}
		} else {
			klog.Infof("skipping comment download for #%d - not updated", pr.GetNumber())
		}

		co := h.PRSummary(pr, comments)
		co.Labels = pr.Labels
		h.seen[co.URL] = co

		if !postFetchMatch(co, fs) {
			klog.V(4).Infof("PR #%d did not pass postFetchMatch with filter: %v", pr.GetNumber(), fs)
			continue
		}

		filtered = append(filtered, co)
	}

	// TODO: Make this only happen when caches are missed
	if err := h.updateSimilarConversations(filtered); err != nil {
		klog.Errorf("update similar: %v", err)
	}

	klog.Infof("%d of %d PR's within %s/%s matched filters:\n%s", len(filtered), len(prs), org, project, toYAML(fs))
	return filtered, nil
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
