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
	"time"

	"github.com/google/go-github/v24/github"
	"k8s.io/klog"
)

// closedPRDays is how old of a closed PR to consider
const closedPRDays = 14

type PRCommentCache struct {
	Time    time.Time
	Content []*github.PullRequestComment
}

type PRSearchCache struct {
	Time    time.Time
	Content []*github.PullRequest
}

func prSearchKey(org, project string, state string, days int) string {
	if days > 0 {
		return fmt.Sprintf("%s-%s-%s-prs-within-%dd", org, project, state, days)
	}
	return fmt.Sprintf("%s-%s-%s-prs", org, project, state)
}

func (h *HubBub) flushPRSearchCache(org string, project string, minAge time.Duration) error {
	klog.Infof("flushPRs older than %s: %s/%s", minAge, org, project)

	keys := []string{
		issueSearchKey(org, project, "open", 0),
		issueSearchKey(org, project, "closed", closedIssueDays),
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

func (h *HubBub) cachedPRs(ctx context.Context, org string, project string, state string, updatedDays int) ([]*github.PullRequest, error) {
	key := prSearchKey(org, project, state, updatedDays)
	if x, ok := h.cache.Get(key); ok {
		klog.V(1).Infof("cache hit: %s", key)
		prs := x.(PRSearchCache)
		return prs.Content, nil
	}
	klog.Infof("cache miss: %s", key)
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
	}

	entry := PRSearchCache{
		Time:    time.Now(),
		Content: []*github.PullRequest{},
	}

	since := time.Now().Add(time.Duration(updatedDays*-24) * time.Hour)
	foundOldest := false

	for {
		klog.Infof("Downloading %s pull requests for %s/%s (page %d)...", state, org, project, opt.Page)
		xprs, resp, err := h.client.PullRequests.List(ctx, org, project, opt)
		if err != nil {
			klog.Errorf("err")
			return xprs, err
		}
		klog.Errorf("...")

		var prs []*github.PullRequest
		// Because PR searches do not support opt.Since
		if updatedDays != 0 {
			for _, pr := range xprs {
				if pr.GetUpdatedAt().Before(since) {
					foundOldest = true
					break
				}
				prs = append(prs, pr)
			}
		} else {
			prs = xprs
		}

		entry.Content = append(entry.Content, prs...)
		if resp.NextPage == 0 || foundOldest {
			break
		}
		opt.Page = resp.NextPage
	}
	h.cache.Set(key, entry, h.maxListAge)
	return entry.Content, nil
}

func (h *HubBub) cachedPRComments(ctx context.Context, org string, project string, num int, minFetchTime time.Time) ([]*github.PullRequestComment, error) {
	key := fmt.Sprintf("%s-%s-%d-pr-comments", org, project, num)
	if x, ok := h.cache.Get(key); ok {
		cs := x.(PRCommentCache)
		if !cs.Time.Before(minFetchTime) {
			klog.V(1).Infof("%s cache hit", key)
			return cs.Content, nil
		}
		klog.Infof("%s near cache hit: %s is earlier than %s", key, cs.Time, minFetchTime)
	}
	klog.Infof("cache miss: %s", key)

	opt := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allComments []*github.PullRequestComment
	for {
		klog.V(2).Infof("Downloading PR comments for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		cs, resp, err := h.client.PullRequests.ListComments(ctx, org, project, num, opt)
		klog.V(2).Infof("Received %d comments", len(cs))
		if err != nil {
			return cs, err
		}
		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	val := PRCommentCache{Time: time.Now(), Content: allComments}
	h.cache.Set(key, val, h.maxEventAge)
	return val.Content, nil
}

func (h *HubBub) PullRequests(ctx context.Context, org string, project string, fs []Filter) ([]*Colloquy, error) {
	fs = openByDefault(fs)

	klog.Infof("Searching %s/%s for PR's matching: %+v", org, project, fs)
	filtered := []*Colloquy{}

	prs, err := h.cachedPRs(ctx, org, project, "open", 0)
	if err != nil {
		return filtered, err
	}
	klog.Infof("open PR count: %d", len(prs))

	cprs, err := h.cachedPRs(ctx, org, project, "closed", closedIssueDays)
	if err != nil {
		return filtered, err
	}
	klog.Infof("closed PR count: %d", len(prs))
	prs = append(prs, cprs...)

	for _, pr := range prs {
		klog.V(4).Infof("Found PR #%d with labels: %+v", pr.GetNumber(), pr.Labels)
		if !matchItem(pr, pr.Labels, fs) {
			klog.V(4).Infof("PR #%d did not pass matchItem :(", pr.GetNumber())
			continue
		}
		comments, err := h.cachedPRComments(ctx, org, project, pr.GetNumber(), pr.GetUpdatedAt())
		if err != nil {
			klog.Errorf("comments: %v", err)
		}

		co := h.PRSummary(pr, comments)
		co.Labels = pr.Labels

		if !matchColloquy(co, fs) {
			klog.V(4).Infof("PR #%d did not pass matchColloquy with filter: %v", pr.GetNumber(), fs)
			continue
		}

		filtered = append(filtered, co)
	}
	return filtered, nil
}

func (h *HubBub) PRSummary(pr *github.PullRequest, cs []*github.PullRequestComment) *Colloquy {
	cl := []CommentLike{}
	reviewed := false
	for _, c := range cs {
		cl = append(cl, CommentLike(c))
		if c.GetPullRequestReviewID() != 0 {
			reviewed = true
		}
	}
	co := h.baseSummary(pr, cl, isMember(pr.GetAuthorAssociation()))
	if reviewed {
		co.Tags = append(co.Tags, "reviewed")
	}

	// Close enough?
	co.ClosedBy = pr.GetMergedBy()
	return co
}
