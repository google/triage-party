// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hubbub

import (
	"fmt"
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/provider"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/logu"
	"github.com/google/triage-party/pkg/persist"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// cachedIssues returns issues, cached if possible
func (h *Engine) cachedIssues(sp models.SearchParams) ([]*models.Issue, time.Time, error) {
	sp.SearchKey = issueSearchKey(sp)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		// Normally the similarity tables are only updated when fresh data is encountered.
		if sp.NewerThan.IsZero() {
			go h.updateSimilarIssues(sp.SearchKey, x.Issues)
		}

		return x.Issues, x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, logu.STime(sp.NewerThan))
	issues, created, err := h.updateIssues(sp)
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
func (h *Engine) updateIssues(sp models.SearchParams) ([]*models.Issue, time.Time, error) {
	start := time.Now()

	sp.IssueListByRepoOptions = models.IssueListByRepoOptions{
		ListOptions: models.ListOptions{PerPage: 100},
		State:       sp.State,
	}

	if sp.UpdateAge != 0 {
		sp.IssueListByRepoOptions.Since = time.Now().Add(-1 * sp.UpdateAge)
	}

	var allIssues []*models.Issue

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
		is, resp, err := pr.IssuesListByRepo(sp)

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

	if err := h.cache.Set(sp.SearchKey, &persist.Thing{Issues: allIssues}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	klog.V(1).Infof("updateIssues %s returning %d issues", sp.SearchKey, len(allIssues))
	return allIssues, start, nil
}

func (h *Engine) cachedIssueComments(sp models.SearchParams) ([]*models.IssueComment, time.Time, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-issue-comments", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.IssueComments, x.Created, nil
	}

	if !sp.Fetch {
		return nil, time.Time{}, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, logu.STime(sp.NewerThan))

	comments, created, err := h.updateIssueComments(sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.IssueComments, x.Created, nil
		}
	}

	return comments, created, err
}

func (h *Engine) updateIssueComments(sp models.SearchParams) ([]*models.IssueComment, time.Time, error) {
	klog.V(1).Infof("Downloading issue comments for %s/%s #%d", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	start := time.Now()

	sp.IssueListCommentsOptions = models.IssueListCommentsOptions{
		ListOptions: models.ListOptions{PerPage: 100},
	}

	var allComments []*models.IssueComment
	for {
		klog.Infof("Downloading comments for %s/%s #%d (page %d)...",
			sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, sp.IssueListCommentsOptions.Page)

		pr := provider.ResolveProviderByHost(sp.Repo.Host)
		cs, resp, err := pr.IssuesListComments(sp)

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

	if err := h.cache.Set(sp.SearchKey, &persist.Thing{IssueComments: allComments}); err != nil {
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

func openByDefault(fs []Filter) []Filter {
	found := false
	for _, f := range fs {
		if f.State != "" {
			found = true
		}
	}
	if !found {
		fs = append(fs, Filter{State: "open"})
	}
	return fs
}

func (h *Engine) createIssueSummary(i *models.Issue, cs []*models.IssueComment, age time.Time) *Conversation {
	cl := []*models.Comment{}
	for _, c := range cs {
		cl = append(cl, models.NewComment(c))
	}

	co := h.createConversation(i, cl, age)
	r := i.GetReactions()
	co.ReactionsTotal += r.GetTotalCount()
	for k, v := range reactions(r) {
		co.Reactions[k] += v
	}
	co.ClosedBy = i.GetClosedBy()

	sort.Slice(co.Tags, func(i, j int) bool { return co.Tags[i].ID < co.Tags[j].ID })
	return co
}

// IssueSummary returns a cached conversation for an issue
func (h *Engine) IssueSummary(i *models.Issue, cs []*models.IssueComment, age time.Time) *Conversation {
	key := i.GetHTMLURL()
	cached, ok := h.seen[key]
	if ok {
		if !cached.Updated.Before(i.GetUpdatedAt()) && cached.CommentsTotal >= len(cs) {
			return h.seen[key]
		}
		klog.Infof("%s in issue cache, but was invalid. Live @ %s (%d comments), cached @ %s (%d comments)  ", i.GetHTMLURL(), i.GetUpdatedAt(), len(cs), cached.Updated, cached.CommentsTotal)
	}

	h.seen[key] = h.createIssueSummary(i, cs, age)
	return h.seen[key]
}

func isBot(u *models.User) bool {
	if u.GetType() == "bot" {
		return true
	}

	if strings.Contains(u.GetBio(), "stale issues") {
		return true
	}

	if strings.HasSuffix(u.GetLogin(), "-bot") || strings.HasSuffix(u.GetLogin(), "-robot") || strings.HasSuffix(u.GetLogin(), "_bot") || strings.HasSuffix(u.GetLogin(), "_robot") {
		return true
	}

	return false
}
