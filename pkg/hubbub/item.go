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
func (h *Engine) conversation(i GitHubItem, cs []*Comment) *Conversation {
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
		Milestone:            i.GetMilestone(),
		Reactions:            map[string]int{},
	}

	// "https://github.com/kubernetes/minikube/issues/7179",
	urlParts := strings.Split(i.GetHTMLURL(), "/")
	co.Organization = urlParts[3]
	co.Project = urlParts[4]

	if i.GetAssignee() != nil {
		co.Assignees = append(co.Assignees, i.GetAssignee())
		co.Tags = append(co.Tags, assignedTag())
	}

	if !authorIsMember {
		co.LatestMemberResponse = i.GetCreatedAt()
	}

	lastQuestion := time.Time{}
	seenCommenters := map[string]bool{}
	seenClosedCommenters := map[string]bool{}
	seenMemberComment := false

	if co.ID == h.debugNumber {
		klog.Errorf("debug conversation: %s", formatStruct(co))
	}

	for _, c := range cs {
		if co.ID == h.debugNumber {
			klog.Errorf("debug conversation comment: %s", formatStruct(c))
		}

		// We don't like their kind around here
		if isBot(c.User) {
			continue
		}

		r := c.Reactions
		if r.GetTotalCount() > 0 {
			co.ReactionsTotal += r.GetTotalCount()
			for k, v := range reactions(r) {
				co.Reactions[k] += v
			}
		}

		if !i.GetClosedAt().IsZero() && c.Created.After(i.GetClosedAt().Add(30*time.Second)) {
			klog.V(1).Infof("#%d: comment after closed on %s: %+v", co.ID, i.GetClosedAt(), c)
			co.ClosedCommentsTotal++
			seenClosedCommenters[*c.User.Login] = true
		}

		if c.User.GetLogin() == i.GetUser().GetLogin() {
			co.LatestAuthorResponse = c.Created
		}

		if c.User.GetLogin() == i.GetAssignee().GetLogin() {
			co.LatestAssigneeResponse = c.Created
		}

		if h.isMember(c.User.GetLogin(), c.AuthorAssoc) && !isBot(c.User) {
			if !co.LatestMemberResponse.After(co.LatestAuthorResponse) && !authorIsMember {
				co.AccumulatedHoldTime += c.Created.Sub(co.LatestAuthorResponse)
			}
			co.LatestMemberResponse = c.Created
			if !seenMemberComment {
				co.Tags = append(co.Tags, commentedTag())
				seenMemberComment = true
			}
		}

		if strings.Contains(c.Body, "?") {
			for _, line := range strings.Split(c.Body, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, ">") {
					continue
				}
				if strings.Contains(line, "?") {
					klog.V(2).Infof("question at %s: %s", c.Created, line)
					lastQuestion = c.Created
				}
			}
		}

		if !seenCommenters[*c.User.Login] {
			co.Commenters = append(co.Commenters, c.User)
			seenCommenters[*c.User.Login] = true
		}
	}

	if co.LatestMemberResponse.After(co.LatestAuthorResponse) {
		klog.V(2).Infof("marking as send: latest member response (%s) is after latest author response (%s)", co.LatestMemberResponse, co.LatestAuthorResponse)
		co.Tags = append(co.Tags, sendTag())
		co.CurrentHoldTime = 0
	} else if !authorIsMember {
		klog.V(2).Infof("marking as recv: author is not member, latest member response (%s) is before latest author response (%s)", co.LatestMemberResponse, co.LatestAuthorResponse)
		co.Tags = append(co.Tags, recvTag())
		co.CurrentHoldTime += time.Since(co.LatestAuthorResponse)
		co.AccumulatedHoldTime += time.Since(co.LatestAuthorResponse)
	}

	if lastQuestion.After(co.LatestMemberResponse) {
		klog.V(2).Infof("marking as recv-q: last question (%s) comes after last member response (%s)", lastQuestion, co.LatestMemberResponse)
		co.Tags = append(co.Tags, recvQTag())
	}

	// Only open milestones for now to keep the query rate lower
	if co.Milestone != nil && co.Milestone.GetState() == "open" {
		co.Tags = append(co.Tags, milestoneTag())
	}

	if !co.LatestAssigneeResponse.IsZero() {
		co.Tags = append(co.Tags, assigneeUpdatedTag())
	}

	if len(cs) > 0 {
		last := cs[len(cs)-1]
		assoc := strings.ToLower(last.AuthorAssoc)
		if assoc == "none" {
			if last.User.GetLogin() == i.GetUser().GetLogin() {
				co.Tags = append(co.Tags, authorLast())
			}
		} else {
			co.Tags = append(co.Tags, assocLast(assoc))
		}
		co.Updated = last.Updated
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
	assignedTo := map[string]bool{}
	for _, a := range co.Assignees {
		assignedTo[a.GetLogin()] = true
	}

	thisRepo := fmt.Sprintf("%s/%s", co.Organization, co.Project)

	for _, t := range timeline {
		if h.debugNumber == co.ID {
			klog.Errorf("debug timeline event: %s", formatStruct(t))
		}

		if t.GetEvent() == "labeled" && t.GetLabel().GetName() == priority {
			klog.V(2).Infof("prioritized at %s", t.GetCreatedAt())
			co.Prioritized = t.GetCreatedAt()
		}

		if t.GetEvent() == "cross-referenced" {
			if assignedTo[t.GetActor().GetLogin()] {
				if t.GetCreatedAt().After(co.LatestAssigneeResponse) {
					co.LatestAssigneeResponse = t.GetCreatedAt()
					co.Tags = append(co.Tags, assigneeUpdatedTag())
				}
			}

			if co.Type == Issue && t.GetSource().GetIssue().IsPullRequest() {
				state := t.GetSource().GetIssue().GetState()
				refRepo := t.GetSource().GetIssue().GetRepository().GetFullName()
				if refRepo != thisRepo {
					continue
				}

				if assignedTo[t.GetActor().GetLogin()] {
					if state == "open" {
						co.Tags = append(co.Tags, assigneeOpenPRTag())
					} else if state == "closed" {
						co.Tags = append(co.Tags, assigneeClosedPRTag())
					}
				} else {
					if state == "open" {
						co.Tags = append(co.Tags, otherOpenPRTag())
					} else if state == "closed" {
						co.Tags = append(co.Tags, otherClosedPRTag())
					}
				}
			}
		}
	}

	co.Tags = dedupTags(co.Tags)
}

func dedupTags(tags []Tag) []Tag {
	deduped := []Tag{}
	seen := map[string]bool{}

	for _, t := range tags {
		if seen[t.ID] {
			continue
		}
		deduped = append(deduped, t)
		seen[t.ID] = true
	}

	return deduped
}

func assignedTag() Tag {
	return Tag{
		ID:          "assigned",
		Description: "Someone is assigned",
	}
}

func commentedTag() Tag {
	return Tag{
		ID:          "commented",
		Description: "A project member has commented on this",
	}
}

func otherOpenPRTag() Tag {
	return Tag{
		ID:          "other-open-pr",
		Description: "Issue has an open cross-referenced PR",
	}
}

func otherClosedPRTag() Tag {
	return Tag{
		ID:          "other-closed-pr",
		Description: "Issue has an closed cross-referenced PR",
	}
}

func assigneeOpenPRTag() Tag {
	return Tag{
		ID:          "assignee-open-pr",
		Description: "Issue has an open cross-referenced PR",
	}
}

func assigneeClosedPRTag() Tag {
	return Tag{
		ID:          "assignee-closed-pr",
		Description: "Issue has an closed cross-referenced PR",
	}
}

func assigneeUpdatedTag() Tag {
	return Tag{
		ID:          "assignee-updated",
		Description: "The assignee has updated the issue",
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

func milestoneTag() Tag {
	return Tag{
		ID:          "milestone",
		Description: "The issue has a milestone associated to it",
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
