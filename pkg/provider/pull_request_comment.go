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

package provider

import (
	"time"
)

// PullRequestComment represents a comment left on a pull request.
type PullRequestComment struct {
	ID                  *int64     `json:"id,omitempty"`
	NodeID              *string    `json:"node_id,omitempty"`
	InReplyTo           *int64     `json:"in_reply_to_id,omitempty"`
	Body                *string    `json:"body,omitempty"`
	Path                *string    `json:"path,omitempty"`
	DiffHunk            *string    `json:"diff_hunk,omitempty"`
	PullRequestReviewID *int64     `json:"pull_request_review_id,omitempty"`
	Position            *int       `json:"position,omitempty"`
	OriginalPosition    *int       `json:"original_position,omitempty"`
	StartLine           *int       `json:"start_line,omitempty"`
	Line                *int       `json:"line,omitempty"`
	OriginalLine        *int       `json:"original_line,omitempty"`
	OriginalStartLine   *int       `json:"original_start_line,omitempty"`
	Side                *string    `json:"side,omitempty"`
	StartSide           *string    `json:"start_side,omitempty"`
	CommitID            *string    `json:"commit_id,omitempty"`
	OriginalCommitID    *string    `json:"original_commit_id,omitempty"`
	User                *User      `json:"user,omitempty"`
	Reactions           *Reactions `json:"reactions,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	// AuthorAssociation is the comment author's relationship to the pull request's repository.
	// Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".
	AuthorAssociation *string `json:"author_association,omitempty"`
	URL               *string `json:"url,omitempty"`
	HTMLURL           *string `json:"html_url,omitempty"`
	PullRequestURL    *string `json:"pull_request_url,omitempty"`
}

// GetAuthorAssociation returns the AuthorAssociation field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetAuthorAssociation() string {
	if p == nil || p.AuthorAssociation == nil {
		return ""
	}
	return *p.AuthorAssociation
}

// GetBody returns the Body field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetBody() string {
	if p == nil || p.Body == nil {
		return ""
	}
	return *p.Body
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetCreatedAt() time.Time {
	if p == nil || p.CreatedAt == nil {
		return time.Time{}
	}
	return *p.CreatedAt
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetHTMLURL() string {
	if p == nil || p.HTMLURL == nil {
		return ""
	}
	return *p.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetID() int64 {
	if p == nil || p.ID == nil {
		return 0
	}
	return *p.ID
}

// GetPullRequestReviewID returns the PullRequestReviewID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetPullRequestReviewID() int64 {
	if p == nil || p.PullRequestReviewID == nil {
		return 0
	}
	return *p.PullRequestReviewID
}

// GetReactions returns the Reactions field.
func (p *PullRequestComment) GetReactions() *Reactions {
	if p == nil {
		return nil
	}
	return p.Reactions
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetUpdatedAt() time.Time {
	if p == nil || p.UpdatedAt == nil {
		return time.Time{}
	}
	return *p.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetURL() string {
	if p == nil || p.URL == nil {
		return ""
	}
	return *p.URL
}

// GetUser returns the User field.
func (p *PullRequestComment) GetUser() *User {
	if p == nil {
		return nil
	}
	return p.User
}

func (i PullRequestComment) String() string {
	return Stringify(i)
}
