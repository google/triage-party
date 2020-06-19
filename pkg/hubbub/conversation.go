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
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/tag"
)

// Issue is a type representing an issue
const Issue = "issue"

// PullRequest is a type representing a PR
const PullRequest = "pull_request"

// Conversation represents a discussion within a GitHub item (issue or PR)
type Conversation struct {
	ID int `json:"id"`

	Organization string `json:"organization"`
	Project      string `json:"project"`

	URL     string       `json:"url"`
	Title   string       `json:"title"`
	Author  *github.User `json:"author"`
	Type    string       `json:"type"`
	State   string       `json:"state"`
	Created time.Time    `json:"created"`

	// Latest comment or event
	Updated time.Time `json:"updated"`

	// Seen is the time we last saw this conversation
	Seen time.Time `json:"seen"`

	// When did this item reach the current priority?
	Prioritized time.Time `json:"prioritized"`

	SelfInflicted bool `json:"self_inflicted"`

	ReviewState string `json:"review_state"`

	LatestAuthorResponse   time.Time `json:"latest_author_response"`
	LatestAssigneeResponse time.Time `json:"latest_assignee_response"`
	LatestMemberResponse   time.Time `json:"latest_member_response"`

	AccumulatedHoldTime time.Duration `json:"accumulated_hold_time"`
	CurrentHoldTime     time.Duration `json:"current_hold_time"`

	Assignees []*github.User  `json:"assignees"`
	Labels    []*github.Label `json:"labels"`

	ReactionsTotal    int            `json:"reactions_total"`
	Reactions         map[string]int `json:"reactions"`
	ReactionsPerMonth float64        `json:"reactions_per_month"`

	Commenters         []*github.User `json:"commenters"`
	LastCommentBody    string         `json:"last_comment_body"`
	LastCommentAuthor  *github.User   `json:"last_comment_author"`
	CommentsTotal      int            `json:"comments_total"`
	CommentersTotal    int            `json:"commenters_total"`
	CommentersPerMonth float64        `json:"commenters_per_month"`

	ClosedCommentsTotal   int          `json:"closed_comments_total"`
	ClosedCommentersTotal int          `json:"closed_commenters_total"`
	ClosedAt              time.Time    `json:"closed_at"`
	ClosedBy              *github.User `json:"closed_by"`

	IssueRefs       []*RelatedConversation `json:"issue_refs"`
	PullRequestRefs []*RelatedConversation `json:"pull_request_refs"`

	Tags []tag.Tag `json:"tags"`

	// Similar issues to this one
	Similar []*RelatedConversation `json:"similar"`

	Milestone *github.Milestone `json:"milestone"`
}

// A subset of Conversation for related items (requires less memory than a Conversation)
type RelatedConversation struct {
	Organization string    `json:"org"`
	Project      string    `json:"project"`
	ID           int       `json:"int"`
	Tags         []tag.Tag `json:"tags"`

	URL         string       `json:"url"`
	Title       string       `json:"title"`
	Author      *github.User `json:"author"`
	Type        string       `json:"type"`
	State       string       `json:"state"`
	Created     time.Time    `json:"created"`
	Updated     time.Time    `json:"updated"`
	Seen        time.Time    `json:"seen"`
	ReviewState string       `json:"review_state"`
}

func makeRelated(c *Conversation) *RelatedConversation {
	return &RelatedConversation{
		Organization: c.Organization,
		Project:      c.Project,
		ID:           c.ID,

		URL:     c.URL,
		Title:   c.Title,
		Author:  c.Author,
		Type:    c.Type,
		State:   c.State,
		Created: c.Created,
		Updated: c.Updated,
		Tags:    c.Tags,
		Seen:    c.Seen,
	}
}
