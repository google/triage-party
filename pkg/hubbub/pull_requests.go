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
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"

	"github.com/google/go-github/v33/github"
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
func (h *Engine) cachedPRs(ctx context.Context, sp provider.SearchParams) ([]*provider.PullRequest, time.Time, error) {
	sp.SearchKey = prSearchKey(sp)
	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		// Normally the similarity tables are only updated when fresh data is encountered.
		if sp.NewerThan.IsZero() {
			go h.updateSimilarPullRequests(sp.SearchKey, x.PullRequests)
		}
		return x.PullRequests, x.Created, nil
	}

	klog.V(1).Infof("cache miss: %s newer than %s", sp.SearchKey, sp.NewerThan)
	prs, created, err := h.updatePRs(ctx, sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.PullRequests, x.Created, nil
		}
	}
	return prs, created, err
}

// updatePRs returns and caches live PR's
func (h *Engine) updatePRs(ctx context.Context, sp provider.SearchParams) ([]*provider.PullRequest, time.Time, error) {
	start := time.Now()
	sp.PullRequestListOptions = provider.PullRequestListOptions{
		ListOptions: provider.ListOptions{PerPage: 100},
		State:       sp.State,
		Sort:        constants.UpdatedSortOption,
		Direction:   constants.DescDirectionOption,
	}
	klog.V(1).Infof("%s PR list opts for %s: %+v", sp.State, sp.SearchKey, sp.PullRequestListOptions)

	foundOldest := false
	var allPRs []*provider.PullRequest
	for {
		if sp.UpdateAge == 0 {
			klog.Infof("Downloading %s pull requests for %s/%s (page %d)...",
				sp.State, sp.Repo.Organization, sp.Repo.Project, sp.PullRequestListOptions.Page)
		} else {
			klog.Infof("Downloading %s pull requests for %s/%s updated within %s (page %d)...",
				sp.State, sp.Repo.Organization, sp.Repo.Project, sp.UpdateAge, sp.PullRequestListOptions.Page)
		}

		pr := h.provider(sp.Repo.Host)
		prs, resp, err := pr.PullRequestsList(ctx, sp)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				klog.Errorf("oh snap! We reached the GitHub search API limit: %v", err)
			}
			return prs, start, err
		}
		h.logRate(resp.Rate)

		for _, pr := range prs {
			// Because PR searches do not support opt.Since
			if sp.UpdateAge != 0 {
				if time.Since(pr.GetUpdatedAt()) > sp.UpdateAge {
					foundOldest = true
					break
				}
			}

			h.updateMtime(pr, pr.GetUpdatedAt())

			allPRs = append(allPRs, pr)
		}

		go h.updateSimilarPullRequests(sp.SearchKey, prs)

		if resp.NextPage == 0 || resp.NextPage == sp.PullRequestListOptions.Page || foundOldest {
			break
		}
		sp.PullRequestListOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{PullRequests: allPRs}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	klog.V(1).Infof("updatePRs %s returning %d PRs", sp.SearchKey, len(allPRs))

	return allPRs, start, nil
}

func (h *Engine) cachedPR(ctx context.Context, sp provider.SearchParams) (*provider.PullRequest, time.Time, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-pr", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.PullRequests[0], x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, sp.NewerThan)
	if !sp.Fetch {
		return nil, time.Time{}, nil
	}

	pr, created, err := h.updatePR(ctx, sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.PullRequests[0], x.Created, nil
		}
	}
	return pr, created, err
}

// pr gets a single PR (not used very often)
func (h *Engine) updatePR(ctx context.Context, sp provider.SearchParams) (*provider.PullRequest, time.Time, error) {
	klog.V(1).Infof("Downloading single PR %s/%s #%d", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	start := time.Now()

	p := h.provider(sp.Repo.Host)
	pr, resp, err := p.PullRequestsGet(ctx, sp)
	if err != nil {
		return pr, start, err
	}

	h.logRate(resp.Rate)
	h.updateMtime(pr, pr.GetUpdatedAt())

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{PullRequests: []*provider.PullRequest{pr}}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	return pr, start, nil
}

func (h *Engine) cachedReviewComments(ctx context.Context, sp provider.SearchParams) ([]*provider.PullRequestComment, time.Time, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-pr-comments", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.PullRequestComments, x.Created, nil
	}

	if !sp.Fetch {
		return nil, time.Time{}, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, sp.NewerThan)
	comments, created, err := h.updateReviewComments(ctx, sp)
	if err != nil {
		klog.Warningf("Retrieving stale results for %s due to error: %v", sp.SearchKey, err)
		x := h.cache.GetNewerThan(sp.SearchKey, time.Time{})
		if x != nil {
			return x.PullRequestComments, x.Created, nil
		}
	}
	return comments, created, err
}

// prComments mixes together code review comments and pull-request comments
// TODO dont work properly for gitlab - issues API sometimes return 404. Needs investigation
func (h *Engine) prComments(ctx context.Context, sp provider.SearchParams) ([]*provider.Comment, time.Time, error) {
	start := time.Now()

	var comments []*provider.Comment
	cs, _, err := h.cachedIssueComments(ctx, sp)
	if err != nil {
		klog.Errorf("pr comments: %v", err)
	}
	for _, c := range cs {
		comments = append(comments, provider.NewComment(c))
	}

	rc, _, err := h.cachedReviewComments(ctx, sp)
	if err != nil {
		klog.Errorf("comments: %v", err)
	}
	for _, c := range rc {
		h.updateMtimeLong(sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, c.GetUpdatedAt())

		nc := provider.NewComment(c)
		nc.ReviewID = c.GetPullRequestReviewID()
		comments = append(comments, nc)
	}

	// Re-sort the mixture of review and issue comments in ascending time order
	sort.Slice(comments, func(i, j int) bool { return comments[j].Created.After(comments[i].Created) })

	if h.debug[sp.IssueNumber] {
		klog.Errorf("debug comments: %s", formatStruct(comments))
	}

	return comments, start, err
}

func (h *Engine) updateReviewComments(ctx context.Context, sp provider.SearchParams) ([]*provider.PullRequestComment, time.Time, error) {
	klog.V(1).Infof("Downloading review comments for %s/%s #%d", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	start := time.Now()

	sp.ListOptions = provider.ListOptions{PerPage: 100}
	var allComments []*provider.PullRequestComment
	for {
		klog.V(2).Infof("Downloading review comments for %s/%s #%d (page %d)...",
			sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, sp.ListOptions.Page)

		p := h.provider(sp.Repo.Host)
		cs, resp, err := p.PullRequestsListComments(ctx, sp)
		if err != nil {
			return cs, start, err
		}

		h.logRate(resp.Rate)

		klog.V(2).Infof("Received %d review comments", len(cs))
		for _, c := range cs {
			h.updateMtimeLong(sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, c.GetUpdatedAt())
		}
		allComments = append(allComments, cs...)
		if resp.NextPage == 0 {
			break
		}
		sp.ListOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{PullRequestComments: allComments}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	return allComments, start, nil
}

func (h *Engine) createPRSummary(ctx context.Context, sp provider.SearchParams, pr *provider.PullRequest, cs []*provider.Comment,
	timeline []*provider.Timeline, reviews []*provider.PullRequestReview) *Conversation {
	co := h.createConversation(pr, cs, sp.Age)
	co.Type = PullRequest
	co.ReviewsTotal = len(reviews)
	co.TimelineTotal = len(timeline)
	h.addEvents(ctx, sp, co, timeline)

	co.ReviewState = reviewState(pr, timeline, reviews)
	co.Tags[reviewStateTag(co.ReviewState)] = true

	if pr.GetDraft() {
		co.Tags[tag.Draft] = true
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()
	if pr.GetMerged() {
		co.ReviewState = Merged
		co.Tags[tag.Merged] = true
	}

	return co
}

func (h *Engine) PRSummary(ctx context.Context, sp provider.SearchParams, pr *provider.PullRequest, cs []*provider.Comment, timeline []*provider.Timeline,
	reviews []*provider.PullRequestReview) *Conversation {
	key := pr.GetHTMLURL()
	cached, ok := h.seen[key]
	if ok {
		if !cached.Seen.Before(h.mtime(pr)) && cached.CommentsSeen >= len(cs) && cached.TimelineTotal >= len(timeline) && cached.ReviewsTotal >= len(reviews) {
			return h.seen[key]
		}
		if cached.CommentsSeen < len(cs) {
			klog.V(2).Infof("%s in issue cache, but is missing comments. Live @ %s (%d comments), cached @ %s (%d comments)  ", pr.GetHTMLURL(), h.mtime(pr), len(cs), cached.Seen, cached.CommentsSeen)
		} else if cached.TimelineTotal < len(timeline) {
			klog.Infof("%s in issue cache, but is missing timeline events. Live @ %s (%d events), cached @ %s (%d events)  ", pr.GetHTMLURL(), h.mtime(pr), len(timeline), cached.Seen, cached.TimelineTotal)
		} else if cached.ReviewsTotal < len(reviews) {
			klog.Infof("%s in issue cache, but is missing reviews. Live @ %s (%d reviews), cached @ %s (%d reviews)  ", pr.GetHTMLURL(), h.mtime(pr), len(reviews), cached.Seen, cached.ReviewsTotal)
		} else {
			klog.Infof("%s in issue cache, but may be missing updated references. Live @ %s (%d comments), cached @ %s (%d comments)  ", pr.GetHTMLURL(), h.mtime(pr), len(cs), cached.Seen, cached.CommentsSeen)
		}
	}

	h.seen[key] = h.createPRSummary(ctx, sp, pr, cs, timeline, reviews)
	return h.seen[key]
}
