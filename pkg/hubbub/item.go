package hubbub

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog/v2"
)

// GitHubItem is an interface that matches both GitHub Issues and PullRequests
type GitHubItem interface {
	GetAssignee() *github.User
	GetAuthorAssociation() string
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
func (h *Engine) conversation(i GitHubItem, cs []CommentLike) *Conversation {
	authorIsMember := false
	if h.isMember(i.GetUser().GetLogin(), i.GetAuthorAssociation()) {
		authorIsMember = true
	}

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
	seenCommenters := map[string]bool{}
	seenClosedCommenters := map[string]bool{}
	seenMemberComment := false

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
		if h.isMember(c.GetUser().GetLogin(), c.GetAuthorAssociation()) && !isBot(c.GetUser()) {
			if !co.LatestMemberResponse.After(co.LatestAuthorResponse) && !authorIsMember {
				co.AccumulatedHoldTime += c.GetCreatedAt().Sub(co.LatestAuthorResponse)
			}
			co.LatestMemberResponse = c.GetCreatedAt()
			if !seenMemberComment {
				co.Tags = append(co.Tags, commentedTag())
				seenMemberComment = true
			}
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
		co.Tags = append(co.Tags, sendTag())
		co.CurrentHoldTime = 0
	} else if !authorIsMember {
		co.Tags = append(co.Tags, recvTag())
		co.CurrentHoldTime += time.Since(co.LatestAuthorResponse)
		co.AccumulatedHoldTime += time.Since(co.LatestAuthorResponse)
	}

	if lastQuestion.After(co.LatestMemberResponse) {
		co.Tags = append(co.Tags, recvQTag())
	}

	if len(cs) > 0 {
		last := cs[len(cs)-1]
		assoc := strings.ToLower(last.GetAuthorAssociation())
		if assoc == "none" {
			if last.GetUser().GetLogin() == i.GetUser().GetLogin() {
				co.Tags = append(co.Tags, authorLast())
			}
		} else {
			co.Tags = append(co.Tags, assocLast(assoc))
		}
		co.Updated = last.GetUpdatedAt()
	}

	if co.State == "closed" {
		co.Tags = append(co.Tags, closedTag())
	}

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

// Return if a user or role should be considered a member
func (h *Engine) isMember(user string, role string) bool {
	if h.members[user] {
		klog.V(3).Infof("%q (%s) is in membership list", user, role)
		return true
	}

	if h.memberRoles[strings.ToLower(role)] {
		klog.V(3).Infof("%q (%s) is in membership role list", user, role)
		return true
	}

	return false
}

// Add events to the conversation summary if useful
func (h *Engine) addEvents(co *Conversation, timeline []*github.Timeline) {
	priority := ""
	for _, l := range co.Labels {
		if strings.HasPrefix(l.GetName(), "priority") {
			klog.V(1).Infof("found priority: %s", l.GetName())
			priority = l.GetName()
			break
		}
	}

	for _, t := range timeline {
		if t.GetEvent() == "labeled" && t.GetLabel().GetName() == priority {
			klog.Infof("prioritized at %s", t.GetCreatedAt())
			co.Prioritized = t.GetCreatedAt()
		}
	}
}

func commentedTag() Tag {
	return Tag{
		ID:          "commented",
		Description: "A project member has commented on this",
	}
}

func sendTag() Tag {
	return Tag{
		ID:          "send",
		Description: "A project member commented more recently than the author",
	}
}

func recvTag() Tag {
	return Tag{
		ID:          "recv",
		Description: "The author commented more recently than a project member",
	}
}

func recvQTag() Tag {
	return Tag{
		ID:          "recv-q",
		Description: "The author has asked a question since the last project member commented",
	}
}

func authorLast() Tag {
	return Tag{
		ID:          "author-last",
		Description: "The last commenter was the original author",
	}
}

func assocLast(role string) Tag {
	return Tag{
		ID:          fmt.Sprintf("%s-last", role),
		Description: fmt.Sprintf("The last commenter was a project %s", role),
	}
}

func closedTag() Tag {
	return Tag{
		ID:          "closed",
		Description: "This item has been closed",
	}
}
