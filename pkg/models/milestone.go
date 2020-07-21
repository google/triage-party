package models

import "time"

// Milestone represents a GitHub repository milestone.
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

// GetClosedAt returns the ClosedAt field if it's non-nil, zero value otherwise.
func (m *Milestone) GetClosedAt() time.Time {
	if m == nil || m.ClosedAt == nil {
		return time.Time{}
	}
	return *m.ClosedAt
}

// GetClosedIssues returns the ClosedIssues field if it's non-nil, zero value otherwise.
func (m *Milestone) GetClosedIssues() int {
	if m == nil || m.ClosedIssues == nil {
		return 0
	}
	return *m.ClosedIssues
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (m *Milestone) GetCreatedAt() time.Time {
	if m == nil || m.CreatedAt == nil {
		return time.Time{}
	}
	return *m.CreatedAt
}

// GetCreator returns the Creator field.
func (m *Milestone) GetCreator() *User {
	if m == nil {
		return nil
	}
	return m.Creator
}

// GetDescription returns the Description field if it's non-nil, zero value otherwise.
func (m *Milestone) GetDescription() string {
	if m == nil || m.Description == nil {
		return ""
	}
	return *m.Description
}

// GetDueOn returns the DueOn field if it's non-nil, zero value otherwise.
func (m *Milestone) GetDueOn() time.Time {
	if m == nil || m.DueOn == nil {
		return time.Time{}
	}
	return *m.DueOn
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (m *Milestone) GetHTMLURL() string {
	if m == nil || m.HTMLURL == nil {
		return ""
	}
	return *m.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (m *Milestone) GetID() int64 {
	if m == nil || m.ID == nil {
		return 0
	}
	return *m.ID
}

// GetLabelsURL returns the LabelsURL field if it's non-nil, zero value otherwise.
func (m *Milestone) GetLabelsURL() string {
	if m == nil || m.LabelsURL == nil {
		return ""
	}
	return *m.LabelsURL
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (m *Milestone) GetNodeID() string {
	if m == nil || m.NodeID == nil {
		return ""
	}
	return *m.NodeID
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

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (m *Milestone) GetUpdatedAt() time.Time {
	if m == nil || m.UpdatedAt == nil {
		return time.Time{}
	}
	return *m.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (m *Milestone) GetURL() string {
	if m == nil || m.URL == nil {
		return ""
	}
	return *m.URL
}
