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

	"github.com/google/go-github/v24/github"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

// closedIssueDays is how old of a closed issue to consider
const closedIssueDays = 30

type IssueCommentCache struct {
	Time    time.Time
	Content []*github.IssueComment
}

type IssueSearchCache struct {
	Time    time.Time
	Content []*github.Issue
}

func issueSearchKey(org string, project string, state string, days int) string {
	if days > 0 {
		return fmt.Sprintf("%s-%s-%s-issues-within-%dd", org, project, state, days)
	}
	return fmt.Sprintf("%s-%s-%s-issues", org, project, state)
}

func (h *HubBub) flushIssueSearchCache(org string, project string, minAge time.Duration) error {
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

func (h *HubBub) cachedIssues(ctx context.Context, org string, project string, state string, updatedDays int) ([]*github.Issue, error) {
	key := issueSearchKey(org, project, state, updatedDays)
	if x, ok := h.cache.Get(key); ok {
		klog.V(1).Infof("cache hit: %s", key)
		is := x.(IssueSearchCache)
		return is.Content, nil
	}
	klog.V(1).Infof("cache miss: %s", key)

	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
	}
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
		entry.Content = append(entry.Content, is...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	h.cache.Set(key, entry, h.maxListAge)
	return entry.Content, nil
}

func (h *HubBub) cachedIssueComments(ctx context.Context, org string, project string, num int, minFetchTime time.Time) ([]*github.IssueComment, error) {
	key := fmt.Sprintf("%s-%s-%d-issue-comments", org, project, num)
	if x, ok := h.cache.Get(key); ok {
		cs := x.(IssueCommentCache)
		if !cs.Time.Before(minFetchTime) {
			return cs.Content, nil
		}
		klog.V(1).Infof("%s near cache hit: %s is earlier than %s", key, cs.Time, minFetchTime)
	}
	klog.V(1).Infof("cache miss: %s", key)

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
	return string(s)
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

func (h *HubBub) Issues(ctx context.Context, org string, project string, fs []Filter) ([]*Colloquy, error) {
	fs = openByDefault(fs)

	klog.Infof("Gathering raw data for %s/%s search:\n%s", org, project, toYAML(fs))
	var filtered []*Colloquy

	is, err := h.cachedIssues(ctx, org, project, "open", 0)
	if err != nil {
		return filtered, err
	}
	klog.Infof("open issue count: %d", len(is))

	cis, err := h.cachedIssues(ctx, org, project, "closed", closedIssueDays)
	if err != nil {
		return filtered, err
	}
	klog.Infof("closed issue count: %d", len(cis))

	is = append(is, cis...)

	member, err := h.cachedOrgMembers(ctx, org)
	if err != nil {
		return filtered, err
	}

	klog.Infof("Found %d raw issues within %s/%s, filtering for:\n%s", len(is), org, project, toYAML(fs))
	for _, i := range is {
		if i.IsPullRequest() {
			continue
		}

		// Inconsistency warning: issues use a list of labels, prs a list of label pointers
		labels := []*github.Label{}
		for _, l := range i.Labels {
			l := l
			labels = append(labels, &l)
		}

		if !matchItem(i, labels, fs) {
			klog.V(1).Infof("#%d did not match item", i.GetNumber())
			continue
		}

		comments, err := h.cachedIssueComments(ctx, org, project, i.GetNumber(), i.GetUpdatedAt())
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		co := h.IssueSummary(i, comments, member[i.User.GetLogin()])
		co.Labels = labels
		if !matchColloquy(co, fs) {
			klog.V(1).Infof("#%d did not match colloquy", i.GetNumber())
			continue
		}

		filtered = append(filtered, co)
	}
	klog.Infof("%d of %d issues within %s/%s matched filters:\n%s", len(filtered), len(is), org, project, toYAML(fs))
	return filtered, nil
}

type IssueLike interface {
	GetAssignee() *github.User
	GetBody() string
	GetComments() int
	GetHTMLURL() string
	GetCreatedAt() time.Time
	GetID() int64
	GetMilestone() *github.Milestone
	GetNumber() int
	GetClosedAt() time.Time
	GetState() string
	GetTitle() string
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *github.User
	String() string
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

// Check if an item matches the filters, and return any comments downloaded
func matchItem(i IssueLike, labels []*github.Label, fs []Filter) bool {
	for _, f := range fs {
		klog.V(2).Infof("%d: %+v", i.GetNumber(), toYAML(f))

		if f.State != "" && f.State != "all" {
			if i.GetState() != f.State {
				klog.V(3).Infof("#%d state is %q, want: %q", i.GetNumber(), i.GetState(), f.State)
				return false
			}
		}

		if f.Updated != "" {
			if ok := matchDuration(i.GetUpdatedAt(), f.Updated); !ok {
				klog.V(2).Infof("#%d update at %s does not meet %s", i.GetNumber(), i.GetUpdatedAt(), f.Updated)
				return false
			}
		}

		if f.Created != "" {
			if ok := matchDuration(i.GetCreatedAt(), f.Created); !ok {
				klog.V(2).Infof("#%d creation at %s does not meet %s", i.GetNumber(), i.GetCreatedAt(), f.Created)
				return false
			}
		}

		if f.LabelRegex() != nil {
			if ok := matchLabel(labels, f.LabelRegex(), f.LabelNegate()); !ok {
				klog.V(2).Infof("#%d labels do not meet %s", i.GetNumber(), f.LabelRegex())
				return false
			}
		}

		if f.Milestone != "" {
			if i.GetMilestone().GetTitle() != f.Milestone {
				klog.V(2).Infof("#%d milestone does not meet %s: %+v", i.GetNumber(), f.Milestone, i.GetMilestone())
				return false
			}
		}

		// This state can be performed without downloading comments
		if f.TagRegex() != nil && f.TagRegex().String() == "assigned" {
			// If assigned and no assignee, fail
			if !f.TagNegate() && i.GetAssignee() == nil {
				return false
			}
			// if !assigned and has assignee, fail
			if f.TagNegate() && i.GetAssignee() != nil {
				return false
			}
		}
	}
	return true
}

func (h *HubBub) baseSummary(i IssueLike, cs []CommentLike, authorIsMember bool) *Colloquy {
	co := &Colloquy{
		ID:                   i.GetNumber(),
		URL:                  i.GetHTMLURL(),
		Author:               i.GetUser(),
		Title:                i.GetTitle(),
		State:                i.GetState(),
		Type:                 Issue,
		Created:              i.GetCreatedAt(),
		CommentsTotal:        i.GetComments(),
		ClosedAt:             i.GetClosedAt(),
		SelfInflicted:        authorIsMember,
		LatestAuthorResponse: i.GetCreatedAt(),
		Milestone:            i.GetMilestone().GetTitle(),
		Reactions:            map[string]int{},
	}

	if i.GetAssignee() != nil {
		co.Assignees = append(co.Assignees, i.GetAssignee())
	}
	if !authorIsMember {
		co.OnHoldSince = i.GetCreatedAt()
		co.LatestMemberResponse = i.GetCreatedAt()
	}

	lastQuestion := time.Time{}
	tags := map[string]bool{}
	seenCommenters := map[string]bool{}
	seenClosedCommenters := map[string]bool{}

	for _, c := range cs {
		// We don't like their kind around here
		if isBot(c.GetUser()) {
			continue
		}

		r := c.GetReactions()
		if r.GetTotalCount() > 0 {
			co.ReactionsTotal += r.GetTotalCount()
			for k, v := range reactions(r) {
				co.Reactions[k] += v
			}
		}

		if !i.GetClosedAt().IsZero() && c.GetCreatedAt().After(i.GetClosedAt().Add(30*time.Second)) {
			klog.V(1).Infof("#%d: comment after closed on %s: %+v", co.ID, i.GetClosedAt(), c)
			co.ClosedCommentsTotal++
			seenClosedCommenters[*c.GetUser().Login] = true
		}

		if c.GetUser().GetLogin() == i.GetUser().GetLogin() {
			co.LatestAuthorResponse = c.GetCreatedAt()
		}
		if isMember(c.GetAuthorAssociation()) && !isBot(c.GetUser()) {
			if !co.LatestMemberResponse.After(co.LatestAuthorResponse) && !authorIsMember {
				co.LatestResponseDelay = c.GetCreatedAt().Sub(co.LatestAuthorResponse)
				co.OnHoldTotal += co.LatestResponseDelay
			}
			co.LatestMemberResponse = c.GetCreatedAt()
			tags["commented"] = true
		}

		if strings.Contains(c.GetBody(), "?") {
			for _, line := range strings.Split(c.GetBody(), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, ">") {
					continue
				}
				if strings.Contains(line, "?") {
					lastQuestion = c.GetCreatedAt()
				}
			}
		}

		if !seenCommenters[*c.GetUser().Login] {
			co.Commenters = append(co.Commenters, c.GetUser())
			seenCommenters[*c.GetUser().Login] = true
		}
	}

	if co.LatestMemberResponse.After(co.LatestAuthorResponse) {
		tags["send"] = true
		co.OnHoldSince = co.LatestMemberResponse
	} else if !authorIsMember {
		tags["recv"] = true
		co.OnHoldSince = co.LatestAuthorResponse
		co.OnHoldTotal += time.Since(co.LatestAuthorResponse)
		co.LatestResponseDelay = time.Since(co.LatestAuthorResponse)
	}

	if lastQuestion.After(co.LatestMemberResponse) {
		tags["recv-q"] = true
	}

	if len(cs) > 0 {
		last := cs[len(cs)-1]
		assoc := strings.ToLower(last.GetAuthorAssociation())
		if assoc == "none" {
			if last.GetUser().GetLogin() == i.GetUser().GetLogin() {
				tags["author-last"] = true
			} else {
				tags["other-last"] = true
			}
		} else {
			tags[fmt.Sprintf("%s-last", assoc)] = true
		}
		co.Updated = last.GetUpdatedAt()
	}

	if co.State == "closed" {
		tags["closed"] = true
	}

	for k := range tags {
		co.Tags = append(co.Tags, k)
	}
	sort.Strings(co.Tags)
	co.CommentersTotal = len(seenCommenters)
	co.ClosedCommentersTotal = len(seenClosedCommenters)

	// Loose, but good enough
	months := time.Since(co.Created).Hours() / 24 / 30
	co.CommentersPerMonth = float64(co.CommentersTotal) / float64(months)
	co.ReactionsPerMonth = float64(co.ReactionsTotal) / float64(months)
	return co
}

func (h *HubBub) IssueSummary(i *github.Issue, cs []*github.IssueComment, authorIsMember bool) *Colloquy {
	cl := []CommentLike{}
	for _, c := range cs {
		cl = append(cl, CommentLike(c))
	}
	co := h.baseSummary(i, cl, authorIsMember)
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
