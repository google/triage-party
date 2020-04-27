package hubbub

import (
	"context"
	"sync"

	"github.com/golang/glog"
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
			glog.Errorf("members: %v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		open, err = h.cachedIssues(ctx, org, project, "open", 0)
		if err != nil {
			glog.Errorf("open issues: %v", err)
			return
		}
		klog.Infof("%s/%s open issue count: %d", org, project, len(open))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		closed, err = h.cachedIssues(ctx, org, project, "closed", closedIssueDays)
		if err != nil {
			glog.Errorf("closed issues: %v", err)
		}
		klog.Infof("%s/%s closed issue count: %d", org, project, len(closed))
	}()

	wg.Wait()

	var is []*github.Issue
	seen := map[string]bool{}

	for _, i := range append(open, closed...) {
		if seen[i.GetURL()] {
			klog.Errorf("unusual: I already saw #%d", i.GetNumber())
			continue
		}
		seen[i.GetURL()] = true
		is = append(is, i)
	}

	var filtered []*Conversation
	klog.Infof("%s/%s aggregate issue count: %d, filtering for:\n%s", org, project, len(is), toYAML(fs))

	for _, i := range is {
		// Inconsistency warning: issues use a list of labels, prs a list of label pointers
		labels := []*github.Label{}
		for _, l := range i.Labels {
			l := l
			labels = append(labels, l)
		}

		if !matchItem(i, labels, fs) {
			klog.V(1).Infof("#%d - %q did not match item filter: %s", i.GetNumber(), i.GetTitle(), toYAML(fs))
			continue
		}

		comments := []*github.IssueComment{}
		if i.GetComments() > 0 {
			klog.Infof("#%d - %q: need comments for final filtering", i.GetNumber(), i.GetTitle())
			comments, err = h.cachedIssueComments(ctx, org, project, i.GetNumber(), i.GetUpdatedAt())
			if err != nil {
				klog.Errorf("comments: %v", err)
			}
		}

		co := h.IssueSummary(i, comments, members[i.User.GetLogin()])
		co.Labels = labels

		if !matchConversation(co, fs) {
			klog.V(1).Infof("#%d - %q did not match conversation filter: %s", i.GetNumber(), i.GetTitle(), toYAML(fs))
			continue
		}

		filtered = append(filtered, co)
	}

	klog.Infof("%d of %d issues within %s/%s matched filters:\n%s", len(filtered), len(is), org, project, toYAML(fs))
	return filtered, nil
}

func (h *Engine) SearchPullRequests(ctx context.Context, org string, project string, fs []Filter) ([]*Conversation, error) {
	fs = openByDefault(fs)

	klog.Infof("Searching %s/%s for PR's matching: %s", org, project, toYAML(fs))
	filtered := []*Conversation{}

	prs, err := h.cachedPRs(ctx, org, project, "open", 0)
	if err != nil {
		return filtered, err
	}
	klog.Infof("open PR count: %d", len(prs))

	cprs, err := h.cachedPRs(ctx, org, project, "closed", closedIssueDays)
	if err != nil {
		return filtered, err
	}
	klog.Infof("closed PR count: %d", len(prs))
	prs = append(prs, cprs...)

	for _, pr := range prs {
		klog.V(4).Infof("Found PR #%d with labels: %+v", pr.GetNumber(), pr.Labels)
		if !matchItem(pr, pr.Labels, fs) {
			klog.V(4).Infof("PR #%d did not pass matchItem :(", pr.GetNumber())
			continue
		}

		comments := []*github.PullRequestComment{}
		if pr.GetComments() > 0 {
			comments, err = h.cachedPRComments(ctx, org, project, pr.GetNumber(), pr.GetUpdatedAt())
			if err != nil {
				klog.Errorf("comments: %v", err)
			}
		}

		co := h.PRSummary(pr, comments)
		co.Labels = pr.Labels

		if !matchConversation(co, fs) {
			klog.V(4).Infof("PR #%d did not pass matchConversation with filter: %v", pr.GetNumber(), fs)
			continue
		}

		filtered = append(filtered, co)
	}
	return filtered, nil
}
