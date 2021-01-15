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

import "time"

type PullRequestReview struct {
	ID             *int64     `json:"id,omitempty"`
	NodeID         *string    `json:"node_id,omitempty"`
	User           *User      `json:"user,omitempty"`
	Body           *string    `json:"body,omitempty"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	CommitID       *string    `json:"commit_id,omitempty"`
	HTMLURL        *string    `json:"html_url,omitempty"`
	PullRequestURL *string    `json:"pull_request_url,omitempty"`
	State          *string    `json:"state,omitempty"`
	// AuthorAssociation is the comment author's relationship to the issue's repository.
	// Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".
	AuthorAssociation *string `json:"author_association,omitempty"`
}

// GetCommitID returns the CommitID field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetCommitID() string {
	if p == nil || p.CommitID == nil {
		return ""
	}
	return *p.CommitID
}

// GetState returns the State field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetState() string {
	if p == nil || p.State == nil {
		return ""
	}
	return *p.State
}

// GetSubmittedAt returns the SubmittedAt field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetSubmittedAt() time.Time {
	if p == nil || p.SubmittedAt == nil {
		return time.Time{}
	}
	return *p.SubmittedAt
}
