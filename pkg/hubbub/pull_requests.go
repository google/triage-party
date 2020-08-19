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
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

const (
	Unreviewed          = "UNREVIEWED"
	NewCommits          = "NEW_COMMITS"
	ChangesRequested    = "CHANGES_REQUESTED"
	Approved            = "APPROVED"
	PushedAfterApproval = "PUSHED_AFTER_APPROVAL"
	Commented           = "COMMENTED"
	Merged              = "MERGED"
	Closed              = "CLOSED"
)

// cachedPRs returns a list of cached PR's if possible
func (h *Engine) cachedPRs(ctx context.Context, org string, project string, state string, updateAge time.Duration, newerThan time.Time) ([]*github.PullRequest, time.Time, error) {
	key := prSearchKey(org, project, state, updateAge)
	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		// Normally the similarity tables are only updated when fresh data is encountered.
		if newerThan.IsZero() {
			go h.updateSimilarPullRequests(key, x.PullRequests)
		}
		return x.PullRequests, x.Created, nil
	}

	klog.V(1).Infof("cache miss: %s newer than %s", key, newerThan)
	prs, created, err := h.updatePRs(ctx, org, project, state, updateAge, key)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", key, err)
		x := h.cache.GetNewerThan(key, time.Time{})
		if x != nil {
			return x.PullRequests, x.Created, nil
		}
	}
	return prs, created, err
}

// updatePRs returns and caches live PR's
func (h *Engine) updatePRs(ctx context.Context, org string, project string, state string, updateAge time.Duration, key string) ([]*github.PullRequest, time.Time, error) {
	start := time.Now()
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
			return prs, start, err
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

			h.updateMtime(pr, pr.GetUpdatedAt())

			allPRs = append(allPRs, pr)
		}

		go h.updateSimilarPullRequests(key, prs)

		if resp.NextPage == 0 || foundOldest {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{PullRequests: allPRs}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	klog.V(1).Infof("updatePRs %s returning %d PRs", key, len(allPRs))

	return allPRs, start, nil
}

func (h *Engine) cachedPR(ctx context.Context, org string, project string, num int, newerThan time.Time, fetch bool) (*github.PullRequest, time.Time, error) {
	key := fmt.Sprintf("%s-%s-%d-pr", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequests[0], x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	if !fetch {
		return nil, time.Time{}, nil
	}

	pr, created, err := h.updatePR(ctx, org, project, num, key)

	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", key, err)
		x := h.cache.GetNewerThan(key, time.Time{})
		if x != nil {
			return x.PullRequests[0], x.Created, nil
		}
	}
	return pr, created, err
}

// pr gets a single PR (not used very often)
func (h *Engine) updatePR(ctx context.Context, org string, project string, num int, key string) (*github.PullRequest, time.Time, error) {
	klog.V(1).Infof("Downloading single PR %s/%s #%d", org, project, num)
	start := time.Now()

	pr, resp, err := h.client.PullRequests.Get(ctx, org, project, num)

	if err != nil {
		return pr, start, err
	}

	h.logRate(resp.Rate)
	h.updateMtime(pr, pr.GetUpdatedAt())

	if err := h.cache.Set(key, &persist.Thing{PullRequests: []*github.PullRequest{pr}}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return pr, start, nil
}

func (h *Engine) cachedReviewComments(ctx context.Context, org string, project string, num int, newerThan time.Time, fetch bool) ([]*github.PullRequestComment, time.Time, error) {
	key := fmt.Sprintf("%s-%s-%d-pr-comments", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequestComments, x.Created, nil
	}

	if !fetch {
		return nil, time.Time{}, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	comments, created, err := h.updateReviewComments(ctx, org, project, num, key)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", key, err)
		x := h.cache.GetNewerThan(key, time.Time{})
		if x != nil {
			return x.PullRequestComments, x.Created, nil
		}
	}
	return comments, created, err
}

// prComments mixes together code review comments and pull-request comments
func (h *Engine) prComments(ctx context.Context, org string, project string, num int, newerThan time.Time, fetch bool) ([]*Comment, time.Time, error) {
	start := time.Now()

	var comments []*Comment
	cs, _, err := h.cachedIssueComments(ctx, org, project, num, newerThan, fetch)
	if err != nil {
		klog.Errorf("pr comments: %v", err)
	}
	for _, c := range cs {
		comments = append(comments, NewComment(c))
	}

	rc, _, err := h.cachedReviewComments(ctx, org, project, num, newerThan, fetch)
	if err != nil {
		klog.Errorf("comments: %v", err)
	}
	for _, c := range rc {
		h.updateMtimeLong(org, project, num, c.GetUpdatedAt())

		nc := NewComment(c)
		nc.ReviewID = c.GetPullRequestReviewID()
		comments = append(comments, nc)
	}

	// Re-sort the mixture of review and issue comments in ascending time order
	sort.Slice(comments, func(i, j int) bool { return comments[j].Created.After(comments[i].Created) })

	if h.debug[num] {
		klog.Errorf("debug comments: %s", formatStruct(comments))
	}

	return comments, start, err
}

func (h *Engine) updateReviewComments(ctx context.Context, org string, project string, num int, key string) ([]*github.PullRequestComment, time.Time, error) {
	klog.V(1).Infof("Downloading review comments for %s/%s #%d", org, project, num)
	start := time.Now()

	opt := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allComments []*github.PullRequestComment
	for {
		klog.V(2).Infof("Downloading review comments for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		cs, resp, err := h.client.PullRequests.ListComments(ctx, org, project, num, opt)

		if err != nil {
			return cs, start, err
		}

		h.logRate(resp.Rate)

		klog.V(2).Infof("Received %d review comments", len(cs))
		for _, c := range cs {
			h.updateMtimeLong(org, project, num, c.GetUpdatedAt())
		}
		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{PullRequestComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allComments, start, nil
}

func (h *Engine) createPRSummary(ctx context.Context, pr *github.PullRequest, cs []*Comment, timeline []*github.Timeline, reviews []*github.PullRequestReview, age time.Time, fetch bool) *Conversation {
	co := h.createConversation(pr, cs, age)
	co.Type = PullRequest
	co.ReviewsTotal = len(reviews)
	co.TimelineTotal = len(timeline)
	h.addEvents(ctx, co, timeline, fetch)

	co.ReviewState = reviewState(pr, timeline, reviews)
	co.Tags = append(co.Tags, reviewStateTag(co.ReviewState))

	if pr.GetDraft() {
		co.Tags = append(co.Tags, tag.Draft)
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()
	if pr.GetMerged() {
		co.ReviewState = Merged
		co.Tags = append(co.Tags, tag.Merged)
	}

	co.Tags = tag.Dedup(co.Tags)
	sort.Slice(co.Tags, func(i, j int) bool { return co.Tags[i].ID < co.Tags[j].ID })
	return co
}

func (h *Engine) PRSummary(ctx context.Context, pr *github.PullRequest, cs []*Comment, timeline []*github.Timeline, reviews []*github.PullRequestReview, age time.Time, fetch bool) *Conversation {
	key := pr.GetHTMLURL()
	cached, ok := h.seen[key]
	if ok {
		if !cached.Updated.Before(pr.GetUpdatedAt()) && cached.CommentsTotal >= len(cs) && cached.TimelineTotal >= len(timeline) && cached.ReviewsTotal >= len(reviews) {
			return h.seen[key]
		}
		klog.Infof("%s in PR cache, but was invalid. Live @ %s (%d comments), cached @ %s (%d comments)  ", pr.GetHTMLURL(), pr.GetUpdatedAt(), len(cs), cached.Updated, cached.CommentsTotal)
	}

	h.seen[key] = h.createPRSummary(ctx, pr, cs, timeline, reviews, age, fetch)
	return h.seen[key]
}
