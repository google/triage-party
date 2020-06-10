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
	"k8s.io/klog/v2"
)

func (h *Engine) cachedTimeline(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.Timeline, error) {
	key := fmt.Sprintf("%s-%s-%d-timeline", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.Timeline, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updateTimeline(ctx, org, project, num, key)
}

func (h *Engine) updateTimeline(ctx context.Context, org string, project string, num int, key string) ([]*github.Timeline, error) {
	klog.Infof("Downloading event timeline for %s/%s #%d", org, project, num)

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
func (h *Engine) addEvents(ctx context.Context, co *Conversation, timeline []*github.Timeline) {
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
		if h.debugNumber == co.ID {
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
					co.Tags = append(co.Tags, assigneeUpdatedTag())
				}
			}

			ri := t.GetSource().GetIssue()
			if t.GetCreatedAt().After(h.latestXref[ri.GetHTMLURL()]) {
				h.latestXref[ri.GetHTMLURL()] = t.GetCreatedAt()
				klog.V(2).Infof("updated xref time for %s: %s", ri.GetHTMLURL(), t.GetCreatedAt())
			}

			if co.Type == Issue && ri.IsPullRequest() {
				refRepo := ri.GetRepository().GetFullName()
				// Filter out PR's that are part of other repositories for now
				if refRepo != thisRepo {
					klog.V(1).Infof("PR#%d is in %s, rather than %s", ri.GetNumber(), refRepo, thisRepo)
					continue
				}

				ref := h.prRef(ctx, ri)
				co.PullRequestRefs = append(co.PullRequestRefs, ref)
				co.Tags = append(co.Tags, Tag{
					ID:          fmt.Sprintf("pr-%s", strings.ToLower(strings.Replace(ref.ReviewState, "_", "-", -1))),
					Description: fmt.Sprintf("Issue has a cross-referenced PR in state %s", co.ReviewState),
				})

			} else {
				co.IssueRefs = append(co.IssueRefs, h.issueRef(t.GetSource().GetIssue()))
			}
		}
	}

	co.Tags = dedupTags(co.Tags)
}

func (h *Engine) prRef(ctx context.Context, pr *github.Issue) *RelatedConversation {
	klog.V(1).Infof("Creating PR reference for #%d", pr.GetNumber())

	co := h.conversation(pr, nil)
	rel := makeRelated(co)

	timeline, err := h.cachedTimeline(ctx, co.Organization, co.Project, pr.GetNumber(), h.timelineDate(pr))
	if err != nil {
		klog.Errorf("timeline: %v", err)
	}

	var reviews []*github.PullRequestReview
	if pr.GetState() != "closed" {
		reviews, _, err = h.cachedReviews(ctx, co.Organization, co.Project, pr.GetNumber(), pr.GetUpdatedAt())
		if err != nil {
			klog.Errorf("reviews: %v", err)
		}
	}

	rel.ReviewState = reviewState(timeline, reviews)
	klog.V(1).Infof("Determined PR #%d to be in review state %q", pr.GetNumber(), rel.ReviewState)
	return rel
}

func (h *Engine) issueRef(i *github.Issue) *RelatedConversation {
	co := h.conversation(i, nil)
	return makeRelated(co)
}

func assigneeUpdatedTag() Tag {
	return Tag{
		ID:          "assignee-updated",
		Description: "The assignee has updated the issue",
	}
}
