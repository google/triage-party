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
	"context"
	"fmt"
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
func (h *Engine) cachedIssues(ctx context.Context, org string, project string, state string, updateAge time.Duration, newerThan time.Time) ([]*github.Issue, error) {
	key := issueSearchKey(org, project, state, updateAge)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.Issues, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, logu.STime(newerThan))
	return h.updateIssues(ctx, org, project, state, updateAge, key)
}

// updateIssues updates the issues in cache
func (h *Engine) updateIssues(ctx context.Context, org string, project string, state string, updateAge time.Duration, key string) ([]*github.Issue, error) {
	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
	}
	klog.V(2).Infof("%s issue list opts for %s: %+v", state, key, opt)

	if updateAge != 0 {
		opt.Since = time.Now().Add(-1 * updateAge)
	}

	var allIssues []*github.Issue

	for {
		if updateAge == 0 {
			klog.Infof("Downloading %s issues for %s/%s (page %d)...", state, org, project, opt.Page)
		} else {
			klog.Infof("Downloading %s issues for %s/%s updated within %s (page %d)...", state, org, project, updateAge, opt.Page)
		}

		is, resp, err := h.client.Issues.ListByRepo(ctx, org, project, opt)

		if _, ok := err.(*github.RateLimitError); ok {
			klog.Errorf("oh snap! I reached the GitHub search API limit: %v", err)
		}

		if err != nil {
			return is, err
		}
		h.logRate(resp.Rate)

		for _, i := range is {
			if i.IsPullRequest() {
				continue
			}

			h.updateSimilarityTables(i.GetTitle(), i.GetHTMLURL())
			allIssues = append(allIssues, i)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{Issues: allIssues}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	klog.V(1).Infof("updateIssues %s returning %d issues", key, len(allIssues))
	return allIssues, nil
}

func (h *Engine) cachedIssueComments(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.IssueComment, error) {
	key := fmt.Sprintf("%s-%s-%d-issue-comments", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.IssueComments, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, logu.STime(newerThan))
	return h.updateIssueComments(ctx, org, project, num, key)
}

func (h *Engine) updateIssueComments(ctx context.Context, org string, project string, num int, key string) ([]*github.IssueComment, error) {
	klog.V(1).Infof("Downloading issue comments for %s/%s #%d", org, project, num)

	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allComments []*github.IssueComment
	for {
		klog.Infof("Downloading comments for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		cs, resp, err := h.client.Issues.ListComments(ctx, org, project, num, opt)
		klog.V(2).Infof("Received %d comments", len(cs))
		klog.V(2).Infof("response: %+v", resp)

		if err != nil {
			return cs, err
		}
		h.logRate(resp.Rate)

		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{IssueComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allComments, nil
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

type CommentLike interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *github.Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *github.User
	String() string
}

func (h *Engine) IssueSummary(i *github.Issue, cs []*github.IssueComment, authorIsMember bool) *Conversation {
	cl := []CommentLike{}
	for _, c := range cs {
		cl = append(cl, CommentLike(c))
	}
	co := h.conversation(i, cl, authorIsMember)
	r := i.GetReactions()
	co.ReactionsTotal += r.GetTotalCount()
	for k, v := range reactions(r) {
		co.Reactions[k] += v
	}
	co.ClosedBy = i.GetClosedBy()

	sort.Slice(co.Tags, func(i, j int) bool { return co.Tags[i].ID < co.Tags[j].ID })
	return co
}

func isBot(u *github.User) bool {
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

// Return if a role is basically a member
func isMember(role string) bool {
	// Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".
	switch role {
	case "COLLABORATOR", "MEMBER", "OWNER":
		return true
	default:
		return false
	}
}
