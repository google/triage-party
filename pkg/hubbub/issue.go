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
	"fmt"
	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"
	"strings"
	"time"

	"context"
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/logu"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// cachedIssues returns issues, cached if possible
func (h *Engine) cachedIssues(ctx context.Context, sp provider.SearchParams) ([]*provider.Issue, time.Time, error) {
	sp.SearchKey = issueSearchKey(sp)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		// Normally the similarity tables are only updated when fresh data is encountered.
		if sp.NewerThan.IsZero() {
			go h.updateSimilarIssues(sp.SearchKey, x.Issues)
		}

		return x.Issues, x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, logu.STime(sp.NewerThan))
	issues, created, err := h.updateIssues(ctx, sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.Issues, x.Created, nil
		}
	}
	return issues, created, err
}

// updateIssues updates the issues in cache
func (h *Engine) updateIssues(ctx context.Context, sp provider.SearchParams) ([]*provider.Issue, time.Time, error) {
	start := time.Now()

	sp.IssueListByRepoOptions = provider.IssueListByRepoOptions{
		ListOptions: provider.ListOptions{PerPage: 100},
		State:       sp.State,
	}

	if sp.UpdateAge != 0 {
		sp.IssueListByRepoOptions.Since = time.Now().Add(-1 * sp.UpdateAge)
	}

	var allIssues []*provider.Issue

	for {
		if sp.UpdateAge == 0 {
			klog.Infof("Downloading %s issues for %s/%s (page %d)...",
				sp.State, sp.Repo.Organization, sp.Repo.Project, sp.IssueListByRepoOptions.Page)
		} else {
			klog.Infof(
				"Downloading %s issues for %s/%s updated within %s (page %d)...",
				sp.State,
				sp.Repo.Organization,
				sp.Repo.Project,
				sp.UpdateAge,
				sp.IssueListByRepoOptions.Page,
			)
		}
		pr := provider.ResolveProviderByHost(sp.Repo.Host)
		is, resp, err := pr.IssuesListByRepo(ctx, sp)

		if _, ok := err.(*github.RateLimitError); ok {
			klog.Errorf("oh snap! I reached the GitHub search API limit: %v", err)
		}

		if err != nil {
			return is, start, err
		}

		h.logRate(resp.Rate)

		for _, i := range is {
			if i.IsPullRequest() {
				continue
			}

			h.updateMtime(i, i.GetUpdatedAt())
			allIssues = append(allIssues, i)
		}

		go h.updateSimilarIssues(sp.SearchKey, is)

		if resp.NextPage == 0 {
			break
		}
		sp.IssueListByRepoOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{Issues: allIssues}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	klog.V(1).Infof("updateIssues %s returning %d issues", sp.SearchKey, len(allIssues))
	return allIssues, start, nil
}

func (h *Engine) cachedIssueComments(ctx context.Context, sp provider.SearchParams) ([]*provider.IssueComment, time.Time, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-issue-comments", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.IssueComments, x.Created, nil
	}

	if !sp.Fetch {
		return nil, time.Time{}, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, logu.STime(sp.NewerThan))

	comments, created, err := h.updateIssueComments(ctx, sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.IssueComments, x.Created, nil
		}
	}

	return comments, created, err
}

func (h *Engine) updateIssueComments(ctx context.Context, sp provider.SearchParams) ([]*provider.IssueComment, time.Time, error) {
	klog.V(1).Infof("Downloading issue comments for %s/%s #%d", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	start := time.Now()

	sp.IssueListCommentsOptions = provider.IssueListCommentsOptions{
		ListOptions: provider.ListOptions{PerPage: 100},
	}

	var allComments []*provider.IssueComment
	for {
		klog.Infof("Downloading comments for %s/%s #%d (page %d)...",
			sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, sp.IssueListCommentsOptions.Page)

		pr := provider.ResolveProviderByHost(sp.Repo.Host)
		cs, resp, err := pr.IssuesListComments(ctx, sp)

		if err != nil {
			return cs, start, err
		}
		h.logRate(resp.Rate)

		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		sp.IssueListCommentsOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{IssueComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	return allComments, start, nil
}

func toYAML(v interface{}) string {
	s, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("yaml err: %v", err)
	}
	return strings.Replace(strings.TrimSpace(string(s)), "\n", "; ", -1)
}

func openByDefault(sp provider.SearchParams) []provider.Filter {
	found := false
	for _, f := range sp.Filters {
		if f.State != "" {
			found = true
		}
	}
	if !found {
		var state string
		if sp.Repo.Host == constants.GitlabProviderHost {
			state = constants.OpenedState
		} else {
			state = constants.OpenState
		}
		sp.Filters = append(sp.Filters, provider.Filter{State: state})
	}
	return sp.Filters
}

func (h *Engine) createIssueSummary(i *provider.Issue, cs []*provider.IssueComment, age time.Time) *Conversation {
	cl := []*provider.Comment{}
	for _, c := range cs {
		cl = append(cl, provider.NewComment(c))
	}

	co := h.createConversation(i, cl, age)
	r := i.GetReactions()
	co.ReactionsTotal += r.GetTotalCount()
	for k, v := range reactions(r) {
		co.Reactions[k] += v
	}
	co.ClosedBy = i.GetClosedBy()

	return co
}

// IssueSummary returns a cached conversation for an issue
func (h *Engine) IssueSummary(i *provider.Issue, cs []*provider.IssueComment, age time.Time) *Conversation {
	key := i.GetHTMLURL()
	cached, ok := h.seen[key]
	if ok {
		minAge := h.mtime(i)
		if !cached.Seen.Before(minAge) && cached.CommentsSeen >= len(cs) {
			return h.seen[key]
		}
		if cached.CommentsSeen < len(cs) {
			klog.V(2).Infof("%s in issue cache, but is missing comments. Live @ %s (%d comments), cached @ %s (%d comments)  ", i.GetHTMLURL(), minAge, len(cs), cached.Seen, cached.CommentsSeen)
		} else {
			klog.Infof("%s in issue cache, but may be missing updated references. Live @ %s (%d comments), cached @ %s (%d comments)  ", i.GetHTMLURL(), minAge, len(cs), cached.Seen, cached.CommentsSeen)
		}
	}

	h.seen[key] = h.createIssueSummary(i, cs, age)
	return h.seen[key]
}

func isBot(u *provider.User) bool {
	if u.GetType() == "bot" {
		klog.V(3).Infof("%s type=bot", u.GetLogin())
		return true
	}

	if strings.Contains(u.GetBio(), "stale issues") {
		klog.V(3).Infof("%s bio=stale", u.GetLogin())
		return true
	}

	if strings.HasSuffix(u.GetLogin(), "-bot") || strings.HasSuffix(u.GetLogin(), "-robot") || strings.HasSuffix(u.GetLogin(), "_bot") || strings.HasSuffix(u.GetLogin(), "_robot") {
		return true
	}

	return false
}
