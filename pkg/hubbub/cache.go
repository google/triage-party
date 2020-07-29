package hubbub

import (
	"fmt"
	"github.com/google/triage-party/pkg/provider"
)

// issueSearchKey is the cache key used for issues
func issueSearchKey(sp provider.SearchParams) string {
	if sp.UpdateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-issues-within-%.1fh", sp.Repo.Organization, sp.Repo.Project, sp.State, sp.UpdateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-issues", sp.Repo.Organization, sp.Repo.Project, sp.State)
}

// prSearchKey is the cache key used for prs
func prSearchKey(sp provider.SearchParams) string {
	if sp.UpdateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-prs-within-%.1fh", sp.Repo.Organization, sp.Repo.Project, sp.State, sp.UpdateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-prs", sp.Repo.Organization, sp.Repo.Project, sp.State)
}
