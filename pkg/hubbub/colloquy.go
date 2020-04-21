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

	"github.com/google/go-github/v24/github"
)

type Colloquy struct {
	ID      int          `json:"id"`
	Hidden  bool         `json:"hidden"`
	URL     string       `json:"url"`
	Title   string       `json:"title"`
	Author  *github.User `json:"author"`
	Type    string       `json:"type"`
	State   string       `json:"state"`
	Created time.Time    `json:"created"`

	// Latest comment or event
	Updated time.Time `json:"updated"`

	SelfInflicted bool `json:"self_inflicted"`

	Mergeable bool `json:"mergeable"`

	LatestAuthorResponse time.Time     `json:"latest_author_response"`
	LatestMemberResponse time.Time     `json:"latest_member_response"`
	LatestResponseDelay  time.Duration `json:"latest_response_delay"`
	OnHoldSince          time.Time     `json:"on_hold_since"`
	OnHoldTotal          time.Duration `json:"on_hold_total"`

	Assignees []*github.User  `json:"assignees"`
	Labels    []*github.Label `json:"labels"`

	ReactionsTotal    int            `json:"reactions_total"`
	Reactions         map[string]int `json:"reactions"`
	ReactionsPerMonth float64        `json:"commenters_per_month"`

	Commenters         []*github.User `json:"commenters"`
	LastCommentBody    string         `json:"last_comment_body"`
	LastCommentAuthor  *github.User   `json:"last_comment_author"`
	CommentsTotal      int            `json:"comments_total"`
	CommentersTotal    int            `json:"commenters_total"`
	CommentersPerMonth float64        `json:"commenters_per_month"`

	ClosedCommentsTotal   int          `json:"closed_comments_total"`
	ClosedCommentersTotal int          `json:"closed_commenters_total"`
	ClosedAt              time.Time    `json:"closed_at`
	ClosedBy              *github.User `json:"closed_by`

	Tags []string `json:"tags"`
	// Similar is a subset to keep memory usage low
	Similar []RelatedColloquy `json:"similar"`

	Milestone string `json:"milestone"`
}

// A subset of Colloquy, for related items.
type RelatedColloquy struct {
	ID      int          `json:"id"`
	URL     string       `json:"url"`
	Title   string       `json:"title"`
	Author  *github.User `json:"author"`
	Type    string       `json:"type"`
	Created time.Time    `json:"created"`
}
