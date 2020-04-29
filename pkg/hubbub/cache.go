package hubbub

import (
	"fmt"
	"time"

	"k8s.io/klog"
)

// Toggle acceptability of stale results, useful for bootstrapping
func (e *Engine) AcceptStaleResults(b bool) {
	klog.V(1).Infof("Setting stale results=%v", b)
	e.acceptStaleResults = b
}

// FlushSearchCache invalidates the in-memory search cache
func (h *Engine) FlushSearchCache(org string, project string, olderThan time.Time) error {
	if h.acceptStaleResults {
		return fmt.Errorf("stale results enabled, refusing to flush")
	}

	h.flushIssueSearchCache(org, project, olderThan)
	h.flushPRSearchCache(org, project, olderThan)
	return nil
}

func (h *Engine) flushIssueSearchCache(org string, project string, olderThan time.Time) {
	klog.Infof("flushIssues older than %s: %s/%s", olderThan, org, project)

	keys := []string{
		issueSearchKey(org, project, "open", 0),
		issueSearchKey(org, project, "closed", closedIssueDays),
	}

	for _, key := range keys {
		if err := h.cache.DeleteOlderThan(key, olderThan); err != nil {
			klog.Warningf("delete %q: %v", key, err)
		}
	}
}

func (h *Engine) flushPRSearchCache(org string, project string, olderThan time.Time) {
	klog.Infof("flushPRs older than %s: %s/%s", olderThan, org, project)

	keys := []string{
		issueSearchKey(org, project, "open", 0),
		issueSearchKey(org, project, "closed", closedPRDays),
	}

	for _, key := range keys {
		if err := h.cache.DeleteOlderThan(key, olderThan); err != nil {
			klog.Warningf("delete %q: %v", key, err)
		}
	}
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
