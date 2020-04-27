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

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

// closedPRDays is how old of a closed PR to consider
const closedPRDays = 14

// cachedPRs returns a list of cached PR's if possible
func (h *Engine) cachedPRs(ctx context.Context, org string, project string, state string, updatedDays int) ([]*github.PullRequest, error) {
	key := prSearchKey(org, project, state, updatedDays)
	if x, ok := h.cache.Get(key); ok {
		klog.V(1).Infof("cache hit: %s", key)
		prs := x.(PRSearchCache)
		return prs.Content, nil
	}

	klog.Infof("cache miss: %s", key)
	return h.updatePRs(ctx, org, project, state, updatedDays, key)
}

// updatePRs returns and caches live PR's
func (h *Engine) updatePRs(ctx context.Context, org string, project string, state string, updatedDays int, key string) ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
	}
	klog.Infof("%s PR list opts for %s: %+v", state, key, opt)

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

		// messy
		for _, pr := range xprs {
			if pr.GetState() != state {
				klog.Errorf("#%d: I asked for state %q, but got PR in %q - open a go-github bug!", pr.GetNumber(), state, pr.GetState())
				continue
			}
		}

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
	klog.Infof("updatePRs %s returning %d PRs", key, len(entry.Content))
	return entry.Content, nil
}

func (h *Engine) cachedPRComments(ctx context.Context, org string, project string, num int, minFetchTime time.Time) ([]*github.PullRequestComment, error) {
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

	return h.updatePRComments(ctx, org, project, num, key)
}

func (h *Engine) updatePRComments(ctx context.Context, org string, project string, num int, key string) ([]*github.PullRequestComment, error) {
	klog.Infof("Downloading PR comments for %s/%s #%d", org, project, num)

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

func (h *Engine) PRSummary(pr *github.PullRequest, cs []*github.PullRequestComment) *Conversation {
	cl := []CommentLike{}
	reviewed := false
	for _, c := range cs {
		cl = append(cl, CommentLike(c))
		if c.GetPullRequestReviewID() != 0 {
			reviewed = true
		}
	}
	co := h.conversation(pr, cl, isMember(pr.GetAuthorAssociation()))
	if reviewed {
		co.Tags = append(co.Tags, "reviewed")
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()

	return co
}
