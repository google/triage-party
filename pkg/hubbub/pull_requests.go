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
	"github.com/google/triage-party/pkg/initcache"
	"k8s.io/klog/v2"
)

// closedPRDays is how old of a closed PR to consider
const closedPRDays = 14

// cachedPRs returns a list of cached PR's if possible
func (h *Engine) cachedPRs(ctx context.Context, org string, project string, state string, updatedDays int, newerThan time.Time) ([]*github.PullRequest, error) {
	key := prSearchKey(org, project, state, updatedDays)
	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequests, nil
	}

	klog.V(1).Infof("cache miss: %s newer than %s", key, newerThan)
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
	klog.V(1).Infof("%s PR list opts for %s: %+v", state, key, opt)

	since := time.Now().Add(time.Duration(updatedDays*-24) * time.Hour)
	foundOldest := false
	var allPRs []*github.PullRequest
	for {
		klog.Infof("Downloading %s pull requests for %s/%s (page %d)...", state, org, project, opt.Page)
		prs, resp, err := h.client.PullRequests.List(ctx, org, project, opt)
		if err != nil {
			klog.Errorf("err")
			return prs, err
		}

		for _, pr := range prs {
			// Because PR searches do not support opt.Since
			if updatedDays != 0 {
				if pr.GetUpdatedAt().Before(since) {
					foundOldest = true
					break
				}
			}
			allPRs = append(allPRs, pr)
		}

		if resp.NextPage == 0 || foundOldest {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &initcache.Hoard{PullRequests: allPRs}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	h.lastItemUpdate = time.Now()
	klog.V(1).Infof("updatePRs %s returning %d PRs", key, len(allPRs))

	return allPRs, nil
}

func (h *Engine) cachedPRComments(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.PullRequestComment, error) {
	key := fmt.Sprintf("%s-%s-%d-pr-comments", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequestComments, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updatePRComments(ctx, org, project, num, key)
}

func (h *Engine) updatePRComments(ctx context.Context, org string, project string, num int, key string) ([]*github.PullRequestComment, error) {
	klog.V(1).Infof("Downloading PR comments for %s/%s #%d", org, project, num)

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

	if err := h.cache.Set(key, &initcache.Hoard{PullRequestComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allComments, nil
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

	if pr.GetDraft() {
		co.Tags = append(co.Tags, "draft")
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()

	return co
}
