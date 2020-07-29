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
