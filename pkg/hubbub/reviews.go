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
	"fmt"
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/provider"
	"strings"
	"time"

	"context"
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

func (h *Engine) cachedReviews(ctx context.Context, sp models.SearchParams) ([]*models.PullRequestReview, time.Time, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-pr-reviews", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.Reviews, x.Created, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", sp.SearchKey, sp.NewerThan)
	if !sp.Fetch {
		return nil, time.Time{}, nil
	}
	return h.updateReviews(ctx, sp)
}

func (h *Engine) updateReviews(ctx context.Context, sp models.SearchParams) ([]*models.PullRequestReview, time.Time, error) {
	klog.V(1).Infof("Downloading reviews for %s/%s #%d", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	start := time.Now()

	sp.ListOptions = models.ListOptions{PerPage: 100}

	var allReviews []*models.PullRequestReview
	for {
		klog.V(2).Infof("Downloading reviews for %s/%s #%d (page %d)...",
			sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, sp.ListOptions.Page)

		p := provider.ResolveProviderByHost(sp.Repo.Host)
		cs, resp, err := p.PullRequestsListReviews(ctx, sp)

		if err != nil {
			return cs, start, err
		}

		h.logRate(resp.Rate)

		allReviews = append(allReviews, cs...)
		if resp.NextPage == 0 {
			break
		}
		sp.ListOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &models.Thing{Reviews: allReviews}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	return allReviews, start, nil
}

// reviewState parses review events to see where an issue was left off
func reviewState(pr models.IItem, timeline []*models.Timeline, reviews []*models.PullRequestReview) string {
	state := Unreviewed

	if len(timeline) == 0 && len(reviews) == 0 {
		klog.Infof("Asked for a review state for PR#%d, but have no input data", pr.GetNumber())
		return Unreviewed
	}

	lastCommitID := ""
	lastPushTime := time.Time{}
	open := true

	for _, t := range timeline {
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

	klog.V(1).Infof("PR #%d has %d reviews, hoping one is for %s ...", pr.GetNumber(), len(reviews), lastCommitID)
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

func reviewStateTag(st string) tag.Tag {
	switch st {
	case Approved:
		return tag.Approved
	case Commented:
		return tag.ReviewedWithComment
	case ChangesRequested:
		return tag.ChangesRequested
	case NewCommits:
		return tag.NewCommits
	case Unreviewed:
		return tag.Unreviewed
	case PushedAfterApproval:
		return tag.PushedAfterApproval
	case Closed:
		return tag.Closed
	case Merged:
		return tag.Merged
	default:
		klog.Errorf("No known tag for: %q", st)
	}
	return tag.Tag{}
}
