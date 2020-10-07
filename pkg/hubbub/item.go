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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"
	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

var (
	// wordRelRefRe parses relative issue references, like "fixes #3402"
	wordRelRefRe = regexp.MustCompile(`\s#(\d+)\b`)

	// puncRelRefRe parses relative issue references, like "fixes #3402."
	puncRelRefRe = regexp.MustCompile(`\s\#(\d+)[\.\!:\?]`)

	// absRefRe parses absolute issue references, like "fixes http://github.com/minikube/issues/432"
	absRefRe = regexp.MustCompile(`https*://github.com/(\w+)/(\w+)/[ip][us]\w+/(\d+)`)

	// codeRe matches code
	codeRe    = regexp.MustCompile("(?s)```.*?```")
	detailsRe = regexp.MustCompile(`(?s)<details>.*</details>`)
)

// createConversation creates a conversation from an issue-like
func (h *Engine) createConversation(i provider.IItem, cs []*provider.Comment, age time.Time) *Conversation {
	klog.Infof("creating conversation for #%d with %d/%d comments (age: %s)", i.GetNumber(), len(cs), i.GetComments(), age)

	authorIsMember := false
	if h.isMember(i.GetUser().GetLogin(), i.GetAuthorAssociation()) {
		authorIsMember = true
	}

	co := &Conversation{
		ID:            i.GetNumber(),
		URL:           i.GetHTMLURL(),
		Author:        i.GetUser(),
		Title:         i.GetTitle(),
		State:         i.GetState(),
		Type:          Issue,
		Seen:          age,
		Created:       i.GetCreatedAt(),
		Updated:       i.GetUpdatedAt(),
		CommentsTotal: i.GetComments(),
		// How many comments were parsed
		CommentsSeen:         len(cs),
		ClosedAt:             i.GetClosedAt(),
		SelfInflicted:        authorIsMember,
		LatestAuthorResponse: i.GetCreatedAt(),
		Milestone:            i.GetMilestone(),
		Reactions:            map[string]int{},
		LastCommentAuthor:    i.GetUser(),
		LastCommentBody:      i.GetBody(),
		Tags:                 map[tag.Tag]bool{},
	}

	if co.CommentsTotal == 0 {
		co.CommentsTotal = len(cs)
	}

	// "https://github.com/kubernetes/minikube/issues/7179",
	urlParts := strings.Split(i.GetHTMLURL(), "/")
	co.Organization = urlParts[3]
	co.Project = urlParts[4]
	h.parseRefs(i.GetBody(), co, i.GetUpdatedAt())

	if i.GetAssignee() != nil {
		co.Assignees = append(co.Assignees, i.GetAssignee())
		co.Tags[tag.Assigned] = true
	}

	if !authorIsMember {
		co.LatestMemberResponse = i.GetCreatedAt()
	}

	lastQuestion := time.Time{}
	seenCommenters := map[string]bool{}
	seenClosedCommenters := map[string]bool{}
	seenMemberComment := false

	if h.debug[co.ID] {
		klog.Errorf("debug conversation: %s", formatStruct(co))
	}

	for _, c := range cs {
		h.parseRefs(c.Body, co, c.Updated)
		if h.debug[co.ID] {
			klog.Errorf("debug conversation comment: %s", formatStruct(c))
		}

		// We don't like their kind around here
		if isBot(c.User) {
			continue
		}

		co.LastCommentBody = c.Body
		co.LastCommentAuthor = c.User

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
				co.Tags[tag.Commented] = true
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
					lastQuestion = c.Created
				}
			}
		}

		if !seenCommenters[*c.User.Login] {
			co.Commenters = append(co.Commenters, c.User)
			seenCommenters[*c.User.Login] = true
		}
	}

	if co.Milestone != nil && co.Milestone.GetState() == "open" {
		co.Tags[tag.OpenMilestone] = true
	}

	if !co.LatestAssigneeResponse.IsZero() {
		co.Tags[tag.AssigneeUpdated] = true
	}

	// Only add these tags if we've seen all the comments
	if len(cs) >= co.CommentsTotal {
		if co.LatestMemberResponse.After(co.LatestAuthorResponse) {
			co.Tags[tag.Send] = true
			co.CurrentHoldTime = 0
		} else if !authorIsMember {
			co.Tags[tag.Recv] = true
			co.CurrentHoldTime += time.Since(co.LatestAuthorResponse)
			co.AccumulatedHoldTime += time.Since(co.LatestAuthorResponse)
		}

		if lastQuestion.After(co.LatestMemberResponse) {
			co.Tags[tag.RecvQ] = true
		}
	}

	if len(cs) > 0 {
		last := cs[len(cs)-1]
		assoc := strings.ToLower(last.AuthorAssoc)
		if assoc == "none" {
			if last.User.GetLogin() == i.GetUser().GetLogin() {
				co.Tags[tag.AuthorLast] = true
			}
		} else {
			co.Tags[tag.RoleLast(assoc)] = true
		}

		if last.Updated.After(co.Updated) {
			co.Updated = last.Updated
		}
	}

	if co.State == constants.ClosedState {
		co.Tags[tag.Closed] = true
	}

	co.CommentersTotal = len(seenCommenters)
	co.ClosedCommentersTotal = len(seenClosedCommenters)

	if co.AccumulatedHoldTime > time.Since(co.Created) {
		panic(fmt.Sprintf("accumulated %s is more than age %s", co.AccumulatedHoldTime, time.Since(co.Created)))
	}

	// Loose, but good enough
	months := time.Since(co.Created).Hours() / 24 / 30
	co.CommentersPerMonth = float64(co.CommentersTotal) / months
	co.ReactionsPerMonth = float64(co.ReactionsTotal) / months

	tagNames := []string{}
	for k := range co.Tags {
		tagNames = append(tagNames, k.ID)
	}

	if len(tagNames) > 0 {
		klog.V(1).Infof("#%d tags based on %d/%d comments: %s", co.ID, co.CommentsSeen, co.CommentsTotal, tagNames)
	}
	return co
}

// Return if a user or role should be considered a member
func (h *Engine) isMember(user string, role string) bool {
	if h.members[user] {
		return true
	}

	if h.memberRoles[strings.ToLower(role)] {
		return true
	}

	klog.V(1).Infof("%s (%s) is not considered a member: members=%s memberRoles=%s", user, role, h.members, h.memberRoles)
	return false
}

// UpdateIssueRefs updates referenced issues within a conversation, adding it if necessary
func (co *Conversation) UpdateIssueRefs(rc *RelatedConversation) {
	for i, ex := range co.IssueRefs {
		if ex.URL == rc.URL {
			if ex.Seen.After(rc.Seen) {
				return
			}
			co.IssueRefs[i] = rc
			return
		}
	}

	co.IssueRefs = append(co.IssueRefs, rc)
}

// UpdatePullRequestRefs updates referenced PR's within a conversation, adding it if necessary
func (co *Conversation) UpdatePullRequestRefs(rc *RelatedConversation) {
	for i, ex := range co.PullRequestRefs {
		if ex.URL == rc.URL {
			if ex.Seen.After(rc.Seen) {
				return
			}
			co.PullRequestRefs[i] = rc
			return
		}
	}

	co.PullRequestRefs = append(co.PullRequestRefs, rc)
}

// parse any references and update mention time
func (h *Engine) parseRefs(text string, co *Conversation, t time.Time) {
	// remove code samples which mention unrelated issues
	text = codeRe.ReplaceAllString(text, "<code></code>")
	text = detailsRe.ReplaceAllString(text, "<details></details>")

	var ms [][]string
	ms = append(ms, wordRelRefRe.FindAllStringSubmatch(text, -1)...)
	ms = append(ms, puncRelRefRe.FindAllStringSubmatch(text, -1)...)

	seen := map[string]bool{}

	for _, m := range ms {
		i, err := strconv.Atoi(m[1])
		if err != nil {
			klog.Errorf("unable to parse int from %s: %v", m[1], err)
			continue
		}

		if i == co.ID {
			continue
		}

		rc := &RelatedConversation{
			Organization: co.Organization,
			Project:      co.Project,
			ID:           i,
			Seen:         t,
		}

		if t.After(h.mtimeRef(rc)) {
			klog.V(1).Infof("%s later referenced #%d at %s: %s", co.URL, i, t, text)
			h.updateMtimeLong(co.Organization, co.Project, i, t)
		}

		if !seen[fmt.Sprintf("%s/%d", rc.Project, rc.ID)] {
			co.UpdateIssueRefs(rc)
		}
		seen[fmt.Sprintf("%s/%d", rc.Project, rc.ID)] = true
	}

	for _, m := range absRefRe.FindAllStringSubmatch(text, -1) {
		org := m[1]
		project := m[2]
		i, err := strconv.Atoi(m[3])
		if err != nil {
			klog.Errorf("unable to parse int from %s: %v", err)
			continue
		}

		if i == co.ID && org == co.Organization && project == co.Project {
			continue
		}

		rc := &RelatedConversation{
			Organization: org,
			Project:      project,
			ID:           i,
			Seen:         t,
		}

		if t.After(h.mtimeRef(rc)) {
			klog.Infof("%s later referenced %s/%s #%d at %s: %s", co.URL, org, project, i, t, text)
			h.updateMtimeLong(org, project, i, t)
		}

		if !seen[fmt.Sprintf("%s/%d", rc.Project, rc.ID)] {
			co.UpdateIssueRefs(rc)
		}
		seen[fmt.Sprintf("%s/%d", rc.Project, rc.ID)] = true
	}
}
