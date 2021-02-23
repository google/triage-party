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

type Milestone struct {
	URL          *string    `json:"url,omitempty"`
	HTMLURL      *string    `json:"html_url,omitempty"`
	LabelsURL    *string    `json:"labels_url,omitempty"`
	ID           *int64     `json:"id,omitempty"`
	Number       *int       `json:"number,omitempty"`
	State        *string    `json:"state,omitempty"`
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Creator      *User      `json:"creator,omitempty"`
	OpenIssues   *int       `json:"open_issues,omitempty"`
	ClosedIssues *int       `json:"closed_issues,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	DueOn        *time.Time `json:"due_on,omitempty"`
	NodeID       *string    `json:"node_id,omitempty"`
}

// GetDueOn returns the DueOn field if it's non-nil, zero value otherwise.
func (m *Milestone) GetDueOn() time.Time {
	if m == nil || m.DueOn == nil {
		return time.Time{}
	}
	return *m.DueOn
}

// GetNumber returns the Number field if it's non-nil, zero value otherwise.
func (m *Milestone) GetNumber() int {
	if m == nil || m.Number == nil {
		return 0
	}
	return *m.Number
}

// GetOpenIssues returns the OpenIssues field if it's non-nil, zero value otherwise.
func (m *Milestone) GetOpenIssues() int {
	if m == nil || m.OpenIssues == nil {
		return 0
	}
	return *m.OpenIssues
}

// GetClosedIssues returns the ClosedIssues field if it's non-nil, zero value otherwise.
func (m *Milestone) GetClosedIssues() int {
	if m == nil || m.ClosedIssues == nil {
		return 0
	}
	return *m.ClosedIssues
}

// GetState returns the State field if it's non-nil, zero value otherwise.
func (m *Milestone) GetState() string {
	if m == nil || m.State == nil {
		return ""
	}
	return *m.State
}

// GetTitle returns the Title field if it's non-nil, zero value otherwise.
func (m *Milestone) GetTitle() string {
	if m == nil || m.Title == nil {
		return ""
	}
	return *m.Title
}
