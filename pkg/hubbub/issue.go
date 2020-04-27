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
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

// closedIssueDays is how old of a closed issue to consider
const closedIssueDays = 14

// cachedIssues returns issues, cached if possible
func (h *Engine) cachedIssues(ctx context.Context, org string, project string, state string, updatedDays int) ([]*github.Issue, error) {
	key := issueSearchKey(org, project, state, updatedDays)
	if x, ok := h.cache.Get(key); ok {
		is := x.(IssueSearchCache)
		return is.Content, nil
	}

	return h.updateIssues(ctx, org, project, state, updatedDays, key)
}

// updateIssues updates the issues in cache
func (h *Engine) updateIssues(ctx context.Context, org string, project string, state string, updatedDays int, key string) ([]*github.Issue, error) {
	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
	}
	klog.Infof("%s issue list opts for %s: %+v", state, key, opt)

	if updatedDays > 0 {
		opt.Since = time.Now().Add(time.Duration(updatedDays*-24) * time.Hour)
	}

	entry := IssueSearchCache{
		Time:    time.Now(),
		Content: []*github.Issue{},
	}

	for {
		klog.Infof("Downloading %s issues for %s/%s (page %d)...", state, org, project, opt.Page)
		is, resp, err := h.client.Issues.ListByRepo(ctx, org, project, opt)
		if err != nil {
			return is, err
		}

		for _, i := range is {
			if i.IsPullRequest() {
				continue
			}

			if i.GetState() != state {
				klog.Errorf("#%d: I asked for state %q, but got issue in %q - open a go-github bug!", i.GetNumber(), state, i.GetState())
				continue
			}

			entry.Content = append(entry.Content, i)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	h.cache.Set(key, entry, h.maxListAge)
	klog.Infof("updateIssues %s returning %d issues", key, len(entry.Content))
	return entry.Content, nil
}

func (h *Engine) cachedIssueComments(ctx context.Context, org string, project string, num int, minFetchTime time.Time) ([]*github.IssueComment, error) {
	key := fmt.Sprintf("%s-%s-%d-issue-comments", org, project, num)

	if x, ok := h.cache.Get(key); ok {
		cs := x.(IssueCommentCache)
		if !cs.Time.Before(minFetchTime) {
			return cs.Content, nil
		}
		klog.V(1).Infof("%s near cache hit: %s is earlier than %s", key, cs.Time, minFetchTime)
	}

	return h.updateIssueComments(ctx, org, project, num, key)
}

func (h *Engine) updateIssueComments(ctx context.Context, org string, project string, num int, key string) ([]*github.IssueComment, error) {
	klog.Infof("Downloading issue comments for %s/%s #%d", org, project, num)

	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allComments []*github.IssueComment
	for {
		klog.V(2).Infof("Downloading comments for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		cs, resp, err := h.client.Issues.ListComments(ctx, org, project, num, opt)
		klog.V(2).Infof("Received %d comments", len(cs))
		klog.V(2).Infof("response: %+v", resp)

		if err != nil {
			return cs, err
		}
		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	val := IssueCommentCache{Time: time.Now(), Content: allComments}
	h.cache.Set(key, val, h.maxEventAge)
	return val.Content, nil
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
	return co
}

func isBot(u *github.User) bool {
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
