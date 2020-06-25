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
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/persist"
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

func (h *Engine) cachedTimeline(ctx context.Context, org string, project string, num int, newerThan time.Time, fetch bool) ([]*github.Timeline, error) {
	key := fmt.Sprintf("%s-%s-%d-timeline", org, project, num)
	klog.V(1).Infof("Need timeline for %s as of %s", key, newerThan)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.Timeline, nil
	}

	klog.Infof("cache miss for %s newer than %s (fetch=%v)", key, newerThan, fetch)
	if !fetch {
		return nil, nil
	}
	return h.updateTimeline(ctx, org, project, num, key)
}

func (h *Engine) updateTimeline(ctx context.Context, org string, project string, num int, key string) ([]*github.Timeline, error) {
	//	klog.Infof("Downloading event timeline for %s/%s #%d", org, project, num)

	opt := &github.ListOptions{
		PerPage: 100,
	}
	var allEvents []*github.Timeline
	for {
		klog.V(2).Infof("Downloading timeline for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		evs, resp, err := h.client.Issues.ListIssueTimeline(ctx, org, project, num, opt)
		if err != nil {
			return nil, err
		}
		h.logRate(resp.Rate)

		for _, ev := range evs {
			h.updateMtimeLong(org, project, num, ev.GetCreatedAt())
		}

		klog.V(2).Infof("Received %d timeline events", len(evs))
		allEvents = append(allEvents, evs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{Timeline: allEvents}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allEvents, nil
}

// Add events to the conversation summary if useful
func (h *Engine) addEvents(ctx context.Context, co *Conversation, timeline []*github.Timeline, fetch bool) {
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
			klog.V(2).Infof("prioritized at %s", t.GetCreatedAt())
			co.Prioritized = t.GetCreatedAt()
		}

		if t.GetEvent() == "cross-referenced" {
			if assignedTo[t.GetActor().GetLogin()] {
				klog.V(1).Infof("cross-referenced by the assignee, updating assigned response")
				if t.GetCreatedAt().After(co.LatestAssigneeResponse) {
					co.LatestAssigneeResponse = t.GetCreatedAt()
					co.Tags = append(co.Tags, tag.AssigneeUpdated)
				}
			}

			ri := t.GetSource().GetIssue()
			klog.Infof("Found xref: #%d -> #%d at %s", co.ID, ri.GetNumber(), t.GetCreatedAt())

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

				klog.Infof("Found cross-referenced PR: #%d, updating PR ref", ri.GetNumber())
				ref := h.prRef(ctx, ri, h.mtimeCo(co), fetch)
				co.PullRequestRefs = append(co.PullRequestRefs, ref)
				refTag := reviewStateTag(ref.ReviewState)
				refTag.ID = fmt.Sprintf("pr-%s", refTag.ID)
				refTag.Desc = fmt.Sprintf("cross-referenced PR: %s", refTag.Desc)
				co.Tags = append(co.Tags, refTag)
			} else {
				co.IssueRefs = append(co.IssueRefs, h.issueRef(t.GetSource().GetIssue(), co.Seen))
			}
		}
	}

	co.Tags = tag.Dedup(co.Tags)
}

func (h *Engine) prRef(ctx context.Context, pr GitHubItem, age time.Time, fetch bool) *RelatedConversation {
	if pr == nil {
		klog.Errorf("PR is nil")
		return nil
	}

	newerThan := age
	if h.mtime(pr).After(newerThan) {
		newerThan = h.mtime(pr)
	}

	if !pr.GetClosedAt().IsZero() {
		newerThan = pr.GetClosedAt()
	}

	klog.V(1).Infof("Creating PR reference for #%d, updated at %s(state=%s)", pr.GetNumber(), pr.GetUpdatedAt(), pr.GetState())

	co := h.conversation(pr, nil, age)
	rel := makeRelated(co)

	timeline, err := h.cachedTimeline(ctx, co.Organization, co.Project, pr.GetNumber(), newerThan, fetch)
	if err != nil {
		klog.Errorf("timeline: %v", err)
	}

	// mtime may have been updated by fetching tthe timeline
	if h.mtime(pr).After(newerThan) {
		newerThan = h.mtime(pr)
	}

	var reviews []*github.PullRequestReview
	if pr.GetState() != "closed" {
		reviews, _, err = h.cachedReviews(ctx, co.Organization, co.Project, pr.GetNumber(), newerThan, fetch)
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

func (h *Engine) updateLinkedPRs(ctx context.Context, parent *Conversation, newerThan time.Time, fetch bool) []*RelatedConversation {
	newRefs := []*RelatedConversation{}

	for _, ref := range parent.PullRequestRefs {
		if h.mtimeRef(ref).After(newerThan) {
			newerThan = h.mtimeRef(ref)
		}
	}

	for _, ref := range parent.PullRequestRefs {
		if newerThan.Before(ref.Seen) || newerThan == ref.Seen {
			newRefs = append(newRefs, ref)
			continue
		}

		klog.V(1).Infof("updating PR ref: %s/%s #%d from %s to %s", ref.Organization, ref.Project, ref.ID, ref.Seen, newerThan)
		pr, age, err := h.cachedPR(ctx, ref.Organization, ref.Project, ref.ID, newerThan, fetch)
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

		newRefs = append(newRefs, h.prRef(ctx, pr, age, fetch))
	}

	return newRefs
}

func (h *Engine) issueRef(i *github.Issue, age time.Time) *RelatedConversation {
	co := h.conversation(i, nil, age)
	return makeRelated(co)
}
