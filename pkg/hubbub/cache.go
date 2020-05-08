package hubbub

import (
	"fmt"
	"time"
)

// issueSearchKey is the cache key used for issues
func issueSearchKey(org string, project string, state string, updateAge time.Duration) string {
	if updateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-issues-within-%.1fh", org, project, state, updateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-issues", org, project, state)
}

// prSearchKey is the cache key used for prs
func prSearchKey(org, project string, state string, updateAge time.Duration) string {
	if updateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-prs-within-%.1fh", org, project, state, updateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-prs", org, project, state)
}
