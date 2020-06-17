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
	"github.com/google/triage-party/pkg/persist"
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
			klog.Infof("Building similarity tables from PR cache (%d items)", len(x.PullRequests))
			for _, pr := range x.PullRequests {
				h.updateSimilarityTables(pr.GetTitle(), pr.GetHTMLURL())
			}
		}
		return x.PullRequests, x.Created, nil
	}

	klog.V(1).Infof("cache miss: %s newer than %s", key, newerThan)
	return h.updatePRs(ctx, org, project, state, updateAge, key)
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

			h.updateMtime(pr)

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

	return allPRs, start, nil
}

func (h *Engine) cachedPR(ctx context.Context, org string, project string, num int, newerThan time.Time) (*github.PullRequest, time.Time, error) {
	key := fmt.Sprintf("%s-%s-%d-pr", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequests[0], x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updatePR(ctx, org, project, num, key)
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
	h.updateMtime(pr)

	if err := h.cache.Set(key, &persist.Thing{PullRequests: []*github.PullRequest{pr}}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return pr, start, nil
}

func (h *Engine) cachedReviewComments(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.PullRequestComment, time.Time, error) {
	key := fmt.Sprintf("%s-%s-%d-pr-comments", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.PullRequestComments, x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updateReviewComments(ctx, org, project, num, key)
}

// prComments mixes together code review comments and pull-request comments
func (h *Engine) prComments(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*Comment, time.Time, error) {
	start := time.Now()

	var comments []*Comment
	cs, _, err := h.cachedIssueComments(ctx, org, project, num, newerThan)
	if err != nil {
		klog.Errorf("pr comments: %v", err)
	}
	for _, c := range cs {
		comments = append(comments, NewComment(c))
	}

	rc, _, err := h.cachedReviewComments(ctx, org, project, num, newerThan)
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

	if num == h.debugNumber {
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

func (h *Engine) cachedReviews(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.PullRequestReview, time.Time, error) {
	key := fmt.Sprintf("%s-%s-%d-pr-reviews", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.Reviews, x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updateReviews(ctx, org, project, num, key)
}

func (h *Engine) updateReviews(ctx context.Context, org string, project string, num int, key string) ([]*github.PullRequestReview, time.Time, error) {
	klog.V(1).Infof("Downloading reviews for %s/%s #%d", org, project, num)
	start := time.Now()

	opt := &github.ListOptions{PerPage: 100}

	var allReviews []*github.PullRequestReview
	for {
		klog.V(2).Infof("Downloading reviews for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		cs, resp, err := h.client.PullRequests.ListReviews(ctx, org, project, num, opt)

		if err != nil {
			return cs, start, err
		}

		h.logRate(resp.Rate)

		klog.V(2).Infof("Received %d reviews", len(cs))
		allReviews = append(allReviews, cs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{Reviews: allReviews}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allReviews, start, nil
}

// reviewState parses review events to see where an issue was left off
func reviewState(pr GitHubItem, timeline []*github.Timeline, reviews []*github.PullRequestReview) string {
	state := Unreviewed
	lastCommitID := ""
	lastPushTime := time.Time{}
	open := true

	for _, t := range timeline {
		klog.V(2).Infof("PR #%d review event: %q at %s", pr.GetNumber(), t.GetEvent(), t.GetCreatedAt())
		if t.GetEvent() == "merged" {
			return Merged
		}

		if t.GetEvent() == "head_ref_force_pushed" {
			// GitHub does not return a commit ID
			lastPushTime = t.GetCreatedAt()
		}

		if t.GetEvent() == "committed" {
			commit := t.GetCommitID()
			if commit == "" && strings.Contains(t.GetURL(), "/commits/") {
				parts := strings.Split(t.GetURL(), "/")
				commit = parts[len(parts)-1]
			}
			lastCommitID = commit
		}

		if t.GetEvent() == "reopened" {
			open = true
		}

		if t.GetEvent() == "closed" {
			open = false
		}
	}

	if !open {
		return Closed
	}

	klog.V(1).Infof("PR #%d has %d reviews, hoping one is for %s ...", pr.GetNumber(), (reviews), lastCommitID)
	lastReview := time.Time{}
	for _, r := range reviews {
		if r.GetCommitID() == lastCommitID || lastCommitID == "" {
			klog.V(1).Infof("found %q review at %s for final commit: %s", r.GetState(), r.GetSubmittedAt(), lastCommitID)
			lastReview = r.GetSubmittedAt()
			state = r.GetState()
		} else {
			klog.V(1).Infof("found %q review at %s for older commit: %s", r.GetState(), r.GetSubmittedAt(), r.GetCommitID())
		}
	}

	if state == Unreviewed && len(reviews) > 0 {
		state = NewCommits
	}

	if state == Approved && lastReview.Before(lastPushTime) {
		state = PushedAfterApproval
	}

	return state
}

func reviewStateTag(st string) Tag {
	switch st {
	case Approved:
		return Tag{ID: "approved", Description: "Last review was an approval"}
	case Commented:
		return Tag{ID: "reviewed-with-comment", Description: "Last review was a comment"}
	case ChangesRequested:
		return Tag{ID: "changes-requested", Description: "Last review was a request for changes"}
	case NewCommits:
		return Tag{ID: "new-commits", Description: "PR has commits since the last review"}
	case Unreviewed:
		return Tag{ID: "unreviewed", Description: "PR has never been reviewed"}
	case PushedAfterApproval:
		return Tag{ID: "pushed-after-approval", Description: "PR was pushed to after approval"}
	case Closed:
		return Tag{ID: "closed", Description: "PR was closed"}
	case Merged:
		return Tag{ID: "merged", Description: "PR was merged"}
	default:
		klog.Errorf("No known tag for: %q", st)
	}
	return Tag{}
}

func (h *Engine) PRSummary(ctx context.Context, pr *github.PullRequest, cs []*Comment, timeline []*github.Timeline, reviews []*github.PullRequestReview, age time.Time) *Conversation {
	co := h.conversation(pr, cs, age)
	co.Type = PullRequest
	h.addEvents(ctx, co, timeline)

	co.ReviewState = reviewState(pr, timeline, reviews)
	co.Tags = append(co.Tags, reviewStateTag(co.ReviewState))

	if co.ReviewState != Unreviewed {
		co.Tags = append(co.Tags, Tag{ID: "reviewed", Description: "PR has been reviewed at least once"})
	}

	if pr.GetDraft() {
		co.Tags = append(co.Tags, Tag{ID: "draft", Description: "Draft PR"})
	}

	// Technically not the same thing, but close enough for me.
	co.ClosedBy = pr.GetMergedBy()
	if pr.GetMerged() {
		co.ReviewState = Merged
		co.Tags = append(co.Tags, Tag{ID: "merged", Description: "PR has been merged"})
	}

	sort.Slice(co.Tags, func(i, j int) bool { return co.Tags[i].ID < co.Tags[j].ID })
	return co
}
