package hubbub

import (
	"fmt"
)

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
