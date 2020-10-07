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
	"github.com/google/triage-party/pkg/provider"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

func (h *Engine) cachedTimeline(ctx context.Context, sp provider.SearchParams) ([]*provider.Timeline, error) {
	sp.SearchKey = fmt.Sprintf("%s-%s-%d-timeline", sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	klog.V(1).Infof("Need timeline for %s as of %s", sp.SearchKey, sp.NewerThan)

	if x := h.cache.GetNewerThan(sp.SearchKey, sp.NewerThan); x != nil {
		return x.Timeline, nil
	}

	klog.Infof("cache miss for %s newer than %s (fetch=%v)", sp.SearchKey, sp.NewerThan, sp.Fetch)
	if !sp.Fetch {
		return nil, nil
	}
	return h.updateTimeline(ctx, sp)
}

func (h *Engine) updateTimeline(ctx context.Context, sp provider.SearchParams) ([]*provider.Timeline, error) {
	//	klog.Infof("Downloading event timeline for %s/%s #%d", org, project, num)

	sp.ListOptions = provider.ListOptions{
		PerPage: 100,
	}
	var allEvents []*provider.Timeline
	for {

		pr := provider.ResolveProviderByHost(sp.Repo.Host)
		evs, resp, err := pr.IssuesListIssueTimeline(ctx, sp)
		if err != nil {
			return nil, err
		}
		h.logRate(resp.Rate)

		for _, ev := range evs {
			h.updateMtimeLong(sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, ev.GetCreatedAt())
		}

		allEvents = append(allEvents, evs...)
		if resp.NextPage == 0 || sp.ListOptions.Page == resp.NextPage {
			break
		}
		sp.ListOptions.Page = resp.NextPage
	}

	if err := h.cache.Set(sp.SearchKey, &provider.Thing{Timeline: allEvents}); err != nil {
		klog.Errorf("set %q failed: %v", sp.SearchKey, err)
	}

	return allEvents, nil
}

// Add events to the conversation summary if useful
func (h *Engine) addEvents(ctx context.Context, sp provider.SearchParams, co *Conversation, timeline []*provider.Timeline) {
	priority := ""
	for _, l := range co.Labels {
		if strings.HasPrefix(l.GetName(), "priority") {
			klog.V(1).Infof("found priority: %s", l.GetName())
			priority = l.GetName()
			break
		}
	}
	assignedTo := map[string]bool{}
	for _, a := range co.Assignees {
		assignedTo[a.GetLogin()] = true
	}

	thisRepo := fmt.Sprintf("%s/%s", co.Organization, co.Project)

	for _, t := range timeline {
		if h.debug[co.ID] {
			klog.Errorf("debug timeline event %q: %s", t.GetEvent(), formatStruct(t))
		}

		if t.GetEvent() == "labeled" && t.GetLabel().GetName() == priority {
			co.Prioritized = t.GetCreatedAt()
		}

		if t.GetEvent() == "cross-referenced" {
			if assignedTo[t.GetActor().GetLogin()] {
				if t.GetCreatedAt().After(co.LatestAssigneeResponse) {
					co.LatestAssigneeResponse = t.GetCreatedAt()
					co.Tags[tag.AssigneeUpdated] = true
				}
			}

			ri := t.GetSource().GetIssue()

			// Push the item timestamps as far forwards as possible for the best possible timeline fetch
			h.updateCoMtime(co, t.GetCreatedAt())
			h.updateCoMtime(co, ri.GetUpdatedAt())
			h.updateMtime(ri, t.GetCreatedAt())
			h.updateMtime(ri, ri.GetUpdatedAt())
			h.updateMtime(ri, co.Updated)

			if co.Type == Issue && ri.IsPullRequest() {
				refRepo := ri.GetRepository().GetFullName()
				// Filter out PR's that are part of other repositories for now
				if refRepo != thisRepo {
					klog.V(1).Infof("PR#%d is in %s, rather than %s", ri.GetNumber(), refRepo, thisRepo)
					continue
				}

				klog.V(1).Infof("Found cross-referenced PR: #%d, updating PR ref", ri.GetNumber())

				sp.Age = h.mtimeCo(co)

				ref := h.prRef(ctx, sp, ri)
				co.UpdatePullRequestRefs(ref)
				refTag := reviewStateTag(ref.ReviewState)
				refTag.ID = fmt.Sprintf("pr-%s", refTag.ID)
				refTag.Desc = fmt.Sprintf("cross-referenced PR: %s", refTag.Desc)
				co.Tags[refTag] = true
			} else {
				co.UpdateIssueRefs(h.issueRef(t.GetSource().GetIssue(), co.Seen))
			}
		}
	}
}

func (h *Engine) prRef(ctx context.Context, sp provider.SearchParams, pr provider.IItem) *RelatedConversation {
	if pr == nil {
		klog.Errorf("PR is nil")
		return nil
	}

	sp.NewerThan = sp.Age
	if h.mtime(pr).After(sp.NewerThan) {
		sp.NewerThan = h.mtime(pr)
	}

	if !pr.GetClosedAt().IsZero() {
		sp.NewerThan = pr.GetClosedAt()
	}

	klog.V(1).Infof("Creating PR reference for #%d, updated at %s(state=%s)", pr.GetNumber(), pr.GetUpdatedAt(), pr.GetState())

	co := h.createConversation(pr, nil, sp.Age)
	rel := makeRelated(co)

	sp.Repo.Organization = co.Organization
	sp.Repo.Project = co.Project
	sp.IssueNumber = pr.GetNumber()

	timeline, err := h.cachedTimeline(ctx, sp)
	if err != nil {
		klog.Errorf("timeline: %v", err)
	}

	// mtime may have been updated by fetching tthe timeline
	if h.mtime(pr).After(sp.NewerThan) {
		sp.NewerThan = h.mtime(pr)
	}

	var reviews []*provider.PullRequestReview
	if pr.GetState() != "closed" {
		reviews, _, err = h.cachedReviews(ctx, sp)
		if err != nil {
			klog.Errorf("reviews: %v", err)
		}
	} else {
		klog.V(1).Infof("PR #%d is closed, won't fetch review state", pr.GetNumber())
	}

	rel.ReviewState = reviewState(pr, timeline, reviews)
	klog.V(1).Infof("Determined PR #%d to be in review state %q", pr.GetNumber(), rel.ReviewState)
	return rel
}

func (h *Engine) updateLinkedPRs(ctx context.Context, sp provider.SearchParams, parent *Conversation) []*RelatedConversation {
	newRefs := []*RelatedConversation{}

	for _, ref := range parent.PullRequestRefs {
		if h.mtimeRef(ref).After(sp.NewerThan) {
			sp.NewerThan = h.mtimeRef(ref)
		}
	}

	for _, ref := range parent.PullRequestRefs {
		if sp.NewerThan.Before(ref.Seen) || sp.NewerThan == ref.Seen {
			newRefs = append(newRefs, ref)
			continue
		}

		klog.V(1).Infof("updating PR ref: %s/%s #%d from %s to %s",
			ref.Organization, ref.Project, ref.ID, ref.Seen, sp.NewerThan)

		sp.Repo.Organization = ref.Organization
		sp.Repo.Project = ref.Project
		sp.IssueNumber = ref.ID

		pr, age, err := h.cachedPR(ctx, sp)
		if err != nil {
			klog.Errorf("error updating cached PR: %v", err)
			newRefs = append(newRefs, ref)
			continue
		}
		// unable to fetch
		if pr == nil {
			klog.Warningf("Unable to update PR ref for %s/%s #%d (data not yet available)", ref.Organization, ref.Project, ref.ID)
			newRefs = append(newRefs, ref)
			continue
		}

		sp.Age = age

		newRefs = append(newRefs, h.prRef(ctx, sp, pr))
	}

	return newRefs
}

func (h *Engine) issueRef(i *provider.Issue, age time.Time) *RelatedConversation {
	co := h.createConversation(i, nil, age)
	return makeRelated(co)
}
