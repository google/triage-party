package hubbub

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

// IssueLike is an interface that matches both GitHub Issues and PullRequests
type IssueLike interface {
	GetAssignee() *github.User
	GetBody() string
	GetComments() int
	GetHTMLURL() string
	GetCreatedAt() time.Time
	GetID() int64
	GetMilestone() *github.Milestone
	GetNumber() int
	GetClosedAt() time.Time
	GetState() string
	GetTitle() string
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *github.User
	String() string
}

// conversation creates a conversation from an issue-like
func (h *Engine) conversation(i IssueLike, cs []CommentLike, authorIsMember bool) *Conversation {
	co := &Conversation{
		ID:                   i.GetNumber(),
		URL:                  i.GetHTMLURL(),
		Author:               i.GetUser(),
		Title:                i.GetTitle(),
		State:                i.GetState(),
		Type:                 Issue,
		Created:              i.GetCreatedAt(),
		CommentsTotal:        i.GetComments(),
		ClosedAt:             i.GetClosedAt(),
		SelfInflicted:        authorIsMember,
		LatestAuthorResponse: i.GetCreatedAt(),
		Milestone:            i.GetMilestone().GetTitle(),
		Reactions:            map[string]int{},
	}

	if i.GetAssignee() != nil {
		co.Assignees = append(co.Assignees, i.GetAssignee())
	}

	if !authorIsMember {
		co.LatestMemberResponse = i.GetCreatedAt()
	}

	lastQuestion := time.Time{}
	tags := map[string]bool{}
	seenCommenters := map[string]bool{}
	seenClosedCommenters := map[string]bool{}

	for _, c := range cs {
		// We don't like their kind around here
		if isBot(c.GetUser()) {
			continue
		}

		r := c.GetReactions()
		if r.GetTotalCount() > 0 {
			co.ReactionsTotal += r.GetTotalCount()
			for k, v := range reactions(r) {
				co.Reactions[k] += v
			}
		}

		if !i.GetClosedAt().IsZero() && c.GetCreatedAt().After(i.GetClosedAt().Add(30*time.Second)) {
			klog.V(1).Infof("#%d: comment after closed on %s: %+v", co.ID, i.GetClosedAt(), c)
			co.ClosedCommentsTotal++
			seenClosedCommenters[*c.GetUser().Login] = true
		}

		if c.GetUser().GetLogin() == i.GetUser().GetLogin() {
			co.LatestAuthorResponse = c.GetCreatedAt()
		}
		if isMember(c.GetAuthorAssociation()) && !isBot(c.GetUser()) {
			if !co.LatestMemberResponse.After(co.LatestAuthorResponse) && !authorIsMember {
				co.AccumulatedHoldTime += c.GetCreatedAt().Sub(co.LatestAuthorResponse)
			}
			co.LatestMemberResponse = c.GetCreatedAt()
			tags["commented"] = true
		}

		if strings.Contains(c.GetBody(), "?") {
			for _, line := range strings.Split(c.GetBody(), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, ">") {
					continue
				}
				if strings.Contains(line, "?") {
					lastQuestion = c.GetCreatedAt()
				}
			}
		}

		if !seenCommenters[*c.GetUser().Login] {
			co.Commenters = append(co.Commenters, c.GetUser())
			seenCommenters[*c.GetUser().Login] = true
		}
	}

	if co.LatestMemberResponse.After(co.LatestAuthorResponse) {
		tags["send"] = true
		co.CurrentHoldTime = 0
	} else if !authorIsMember {
		tags["recv"] = true
		co.CurrentHoldTime += time.Since(co.LatestAuthorResponse)
		co.AccumulatedHoldTime += time.Since(co.LatestAuthorResponse)
	}

	if lastQuestion.After(co.LatestMemberResponse) {
		tags["recv-q"] = true
	}

	if len(cs) > 0 {
		last := cs[len(cs)-1]
		assoc := strings.ToLower(last.GetAuthorAssociation())
		if assoc == "none" {
			if last.GetUser().GetLogin() == i.GetUser().GetLogin() {
				tags["author-last"] = true
			} else {
				tags["other-last"] = true
			}
		} else {
			tags[fmt.Sprintf("%s-last", assoc)] = true
		}
		co.Updated = last.GetUpdatedAt()
	}

	if co.State == "closed" {
		tags["closed"] = true
	}

	for k := range tags {
		co.Tags = append(co.Tags, k)
	}
	sort.Strings(co.Tags)
	co.CommentersTotal = len(seenCommenters)
	co.ClosedCommentersTotal = len(seenClosedCommenters)

	if co.AccumulatedHoldTime > time.Since(co.Created) {
		panic(fmt.Sprintf("accumulated %s is more than age %s", co.AccumulatedHoldTime, time.Since(co.Created)))
	}

	// Loose, but good enough
	months := time.Since(co.Created).Hours() / 24 / 30
	co.CommentersPerMonth = float64(co.CommentersTotal) / float64(months)
	co.ReactionsPerMonth = float64(co.ReactionsTotal) / float64(months)
	return co
}

// Check if an item matches the filters, pre-comment fetch
func matchItem(i IssueLike, labels []*github.Label, fs []Filter) bool {
	for _, f := range fs {
		klog.V(2).Infof("matching item #%d against filter: %+v", i.GetNumber(), toYAML(f))

		if f.State != "" && f.State != "all" {
			if i.GetState() != f.State {
				klog.V(3).Infof("#%d state is %q, want: %q", i.GetNumber(), i.GetState(), f.State)
				return false
			}
		}

		if f.Updated != "" {
			if ok := matchDuration(i.GetUpdatedAt(), f.Updated); !ok {
				klog.V(2).Infof("#%d update at %s does not meet %s", i.GetNumber(), i.GetUpdatedAt(), f.Updated)
				return false
			}
		}

		if f.Responded != "" {
			if ok := matchDuration(i.GetUpdatedAt(), f.Responded); !ok {
				klog.V(2).Infof("#%d update at %s does not meet responded %s", i.GetNumber(), i.GetUpdatedAt(), f.Responded)
				return false
			}
		}

		if f.Created != "" {
			if ok := matchDuration(i.GetCreatedAt(), f.Created); !ok {
				klog.V(2).Infof("#%d creation at %s does not meet %s", i.GetNumber(), i.GetCreatedAt(), f.Created)
				return false
			}
		}

		if f.LabelRegex() != nil {
			if ok := matchLabel(labels, f.LabelRegex(), f.LabelNegate()); !ok {
				klog.V(2).Infof("#%d labels do not meet %s", i.GetNumber(), f.LabelRegex())
				return false
			}
		}

		if f.Milestone != "" {
			if i.GetMilestone().GetTitle() != f.Milestone {
				klog.V(2).Infof("#%d milestone does not meet %s: %+v", i.GetNumber(), f.Milestone, i.GetMilestone())
				return false
			}
		}

		// TODO: comment math

		// This state can be performed without downloading comments
		if f.TagRegex() != nil && f.TagRegex().String() == "assigned" {
			// If assigned and no assignee, fail
			if !f.TagNegate() && i.GetAssignee() == nil {
				return false
			}
			// if !assigned and has assignee, fail
			if f.TagNegate() && i.GetAssignee() != nil {
				return false
			}
		}

	}
	return true
}
