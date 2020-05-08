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
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/persist"
	"k8s.io/klog/v2"
)

// cachedPRs returns a list of cached PR's if possible
func (h *Engine) cachedPRs(ctx context.Context, org string, project string, state string, updateAge time.Duration, newerThan time.Time) ([]*github.PullRequest, error) {
	key := prSearchKey(org, project, state, updateAge)
	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequests, nil
	}

	klog.V(1).Infof("cache miss: %s newer than %s", key, newerThan)
	return h.updatePRs(ctx, org, project, state, updateAge, key)
}

// updatePRs returns and caches live PR's
func (h *Engine) updatePRs(ctx context.Context, org string, project string, state string, updateAge time.Duration, key string) ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
	}
	klog.V(1).Infof("%s PR list opts for %s: %+v", state, key, opt)

	foundOldest := false
	var allPRs []*github.PullRequest
	for {
		if updateAge == 0 {
			klog.Infof("Downloading %s pull requests for %s/%s (page %d)...", state, org, project, opt.Page)
		} else {
			klog.Infof("Downloading %s pull requests for %s/%s updated within %s (page %d)...", state, org, project, updateAge, opt.Page)
		}

		prs, resp, err := h.client.PullRequests.List(ctx, org, project, opt)

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				klog.Errorf("oh snap! We reached the GitHub search API limit: %v", err)
			}
			return prs, err
		}
		h.logRate(resp.Rate)

		for _, pr := range prs {
			// Because PR searches do not support opt.Since
			if updateAge != 0 {
				if time.Since(pr.GetUpdatedAt()) > updateAge {
					foundOldest = true
					break
				}
			}

			// TODO: update tables for cached entries too!
			h.updateSimilarityTables(pr.GetTitle(), pr.GetHTMLURL())
			allPRs = append(allPRs, pr)
		}

		if resp.NextPage == 0 || foundOldest {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{PullRequests: allPRs}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

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
		if err != nil {
			return cs, err
		}
		h.logRate(resp.Rate)

		klog.V(2).Infof("Received %d comments", len(cs))
		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{PullRequestComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allComments, nil
}

func (h *Engine) PRSummary(pr *github.PullRequest, cs []*github.PullRequestComment, timeline []*github.Timeline) *Conversation {
	cl := []CommentLike{}
	latestReview := time.Time{}
	for _, c := range cs {
		cl = append(cl, CommentLike(c))
		if c.GetPullRequestReviewID() != 0 {
			latestReview = c.GetCreatedAt()
		}
	}

	co := h.conversation(pr, cl, isMember(pr.GetAuthorAssociation()))
	h.addEvents(co, timeline)

	for _, t := range timeline {
		if t.GetEvent() == "committed" || t.GetEvent() == "head_ref_force_pushed" {
			co.LatestCommit = t.GetCreatedAt()
			if t.GetCreatedAt().After(co.Updated) {
				co.Updated = t.GetCreatedAt()
			}
		}
	}

	if co.LatestCommit.After(co.LatestReview) {
		co.Tags = append(co.Tags, Tag{ID: "new-commits", Description: "PR has commits since the last review"})
	}

	if !latestReview.IsZero() {
		co.LatestReview = latestReview
		co.Tags = append(co.Tags, Tag{ID: "reviewed", Description: "PR has been reviewed at least once"})
	}

	if pr.GetDraft() {
		co.Tags = append(co.Tags, Tag{ID: "draft", Description: "Draft PR"})
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()

	sort.Slice(co.Tags, func(i, j int) bool { return co.Tags[i].ID < co.Tags[j].ID })
	return co
}
