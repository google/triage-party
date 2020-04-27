package hubbub

import (
	"fmt"
	"time"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

// PRCommentCache are cached comments
type PRCommentCache struct {
	Time    time.Time
	Content []*github.PullRequestComment
}

// PRSearchCache are cached PR's
type PRSearchCache struct {
	Time    time.Time
	Content []*github.PullRequest
}

// IssueCommentCache are cached issue comments
type IssueCommentCache struct {
	Time    time.Time
	Content []*github.IssueComment
}

// IssueSearchCache are cached issues
type IssueSearchCache struct {
	Time    time.Time
	Content []*github.Issue
}

// FlushSearchCache invalidates the in-memory search cache
func (h *Engine) FlushSearchCache(org string, project string, minAge time.Duration) error {
	if err := h.flushIssueSearchCache(org, project, minAge); err != nil {
		return fmt.Errorf("issues: %v", err)
	}
	if err := h.flushPRSearchCache(org, project, minAge); err != nil {
		return fmt.Errorf("prs: %v", err)
	}
	return nil
}

func (h *Engine) flushIssueSearchCache(org string, project string, minAge time.Duration) error {
	klog.Infof("flushIssues older than %s: %s/%s", minAge, org, project)

	keys := []string{
		issueSearchKey(org, project, "open", 0),
		issueSearchKey(org, project, "closed", closedIssueDays),
	}

	for _, key := range keys {
		x, ok := h.cache.Get(key)
		if !ok {
			return fmt.Errorf("no such key: %v", key)
		}
		is := x.(IssueSearchCache)
		if time.Since(is.Time) < minAge {
			return fmt.Errorf("%s not old enough: %v", key, is.Time)
		}
		klog.Infof("Flushing %s", key)
		h.cache.Delete(key)
	}
	return nil
}

func (h *Engine) flushPRSearchCache(org string, project string, minAge time.Duration) error {
	klog.Infof("flushPRs older than %s: %s/%s", minAge, org, project)

	keys := []string{
		issueSearchKey(org, project, "open", 0),
		issueSearchKey(org, project, "closed", closedPRDays),
	}

	for _, key := range keys {
		x, ok := h.cache.Get(key)
		if !ok {
			return fmt.Errorf("no such key: %v", key)
		}
		is := x.(PRSearchCache)
		if time.Since(is.Time) < minAge {
			return fmt.Errorf("%s not old enough: %v", key, is.Time)
		}
		klog.Infof("Flushing %s", key)
		h.cache.Delete(key)
	}
	return nil
}

// issueSearchKey is the cache key used for issues
func issueSearchKey(org string, project string, state string, days int) string {
	if days > 0 {
		return fmt.Sprintf("%s-%s-%s-issues-within-%dd", org, project, state, days)
	}
	return fmt.Sprintf("%s-%s-%s-issues", org, project, state)
}

// prSearchKey is the cache key used for prs
func prSearchKey(org, project string, state string, days int) string {
	if days > 0 {
		return fmt.Sprintf("%s-%s-%s-prs-within-%dd", org, project, state, days)
	}
	return fmt.Sprintf("%s-%s-%s-prs", org, project, state)
}
