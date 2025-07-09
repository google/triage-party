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

type Issue struct {
	ID                *int64            `json:"id,omitempty"`
	Number            *int              `json:"number,omitempty"`
	State             *string           `json:"state,omitempty"`
	Locked            *bool             `json:"locked,omitempty"`
	Title             *string           `json:"title,omitempty"`
	Body              *string           `json:"body,omitempty"`
	AuthorAssociation *string           `json:"author_association,omitempty"`
	User              *User             `json:"user,omitempty"`
	Labels            []*Label          `json:"labels,omitempty"`
	Assignee          *User             `json:"assignee,omitempty"`
	Comments          *int              `json:"comments,omitempty"`
	ClosedAt          *time.Time        `json:"closed_at,omitempty"`
	CreatedAt         *time.Time        `json:"created_at,omitempty"`
	UpdatedAt         *time.Time        `json:"updated_at,omitempty"`
	ClosedBy          *User             `json:"closed_by,omitempty"`
	URL               *string           `json:"url,omitempty"`
	HTMLURL           *string           `json:"html_url,omitempty"`
	CommentsURL       *string           `json:"comments_url,omitempty"`
	EventsURL         *string           `json:"events_url,omitempty"`
	LabelsURL         *string           `json:"labels_url,omitempty"`
	RepositoryURL     *string           `json:"repository_url,omitempty"`
	Milestone         *Milestone        `json:"milestone,omitempty"`
	PullRequestLinks  *PullRequestLinks `json:"pull_request,omitempty"`
	Repository        *Repository       `json:"repository,omitempty"`
	Reactions         *Reactions        `json:"reactions,omitempty"`
	Assignees         []*User           `json:"assignees,omitempty"`
	NodeID            *string           `json:"node_id,omitempty"`

	// ActiveLockReason is populated only when LockReason is provided while locking the issue.
	// Possible values are: "off-topic", "too heated", "resolved", and "spam".
	ActiveLockReason *string `json:"active_lock_reason,omitempty"`
}

// GetAssignee returns the Assignee field.
func (i *Issue) GetAssignee() *User {
	if i == nil {
		return nil
	}
	return i.Assignee
}

// GetAssignees returns the Assignee field.
func (i *Issue) GetAssignees() []*User {
	if i == nil {
		return nil
	}
	return i.Assignees
}

// GetAuthorAssociation returns the AuthorAssociation field if it's non-nil, zero value otherwise.
func (i *Issue) GetAuthorAssociation() string {
	if i == nil || i.AuthorAssociation == nil {
		return ""
	}
	return *i.AuthorAssociation
}

// GetBody returns the Body field if it's non-nil, zero value otherwise.
func (i *Issue) GetBody() string {
	if i == nil || i.Body == nil {
		return ""
	}
	return *i.Body
}

// GetClosedAt returns the ClosedAt field if it's non-nil, zero value otherwise.
func (i *Issue) GetClosedAt() time.Time {
	if i == nil || i.ClosedAt == nil {
		return time.Time{}
	}
	return *i.ClosedAt
}

// GetClosedBy returns the ClosedBy field.
func (i *Issue) GetClosedBy() *User {
	if i == nil {
		return nil
	}
	return i.ClosedBy
}

// GetComments returns the Comments field if it's non-nil, zero value otherwise.
func (i *Issue) GetComments() int {
	if i == nil || i.Comments == nil {
		return 0
	}
	return *i.Comments
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (i *Issue) GetCreatedAt() time.Time {
	if i == nil || i.CreatedAt == nil {
		return time.Time{}
	}
	return *i.CreatedAt
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (i *Issue) GetHTMLURL() string {
	if i == nil || i.HTMLURL == nil {
		return ""
	}
	return *i.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (i *Issue) GetID() int64 {
	if i == nil || i.ID == nil {
		return 0
	}
	return *i.ID
}

// GetMilestone returns the Milestone field.
func (i *Issue) GetMilestone() *Milestone {
	if i == nil {
		return nil
	}
	return i.Milestone
}

// GetNumber returns the Number field if it's non-nil, zero value otherwise.
func (i *Issue) GetNumber() int {
	if i == nil || i.Number == nil {
		return 0
	}
	return *i.Number
}

// GetReactions returns the Reactions field.
func (i *Issue) GetReactions() *Reactions {
	if i == nil {
		return nil
	}
	return i.Reactions
}

// GetRepository returns the Repository field.
func (i *Issue) GetRepository() *Repository {
	if i == nil {
		return nil
	}
	return i.Repository
}

// GetState returns the State field if it's non-nil, zero value otherwise.
func (i *Issue) GetState() string {
	if i == nil || i.State == nil {
		return ""
	}
	return *i.State
}

// GetTitle returns the Title field if it's non-nil, zero value otherwise.
func (i *Issue) GetTitle() string {
	if i == nil || i.Title == nil {
		return ""
	}
	return *i.Title
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (i *Issue) GetUpdatedAt() time.Time {
	if i == nil || i.UpdatedAt == nil {
		return time.Time{}
	}
	return *i.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (i *Issue) GetURL() string {
	if i == nil || i.URL == nil {
		return ""
	}
	return *i.URL
}

// GetUser returns the User field.
func (i *Issue) GetUser() *User {
	if i == nil {
		return nil
	}
	return i.User
}

func (i Issue) String() string {
	return Stringify(i)
}

func (i Issue) IsPullRequest() bool {
	return i.PullRequestLinks != nil
}
