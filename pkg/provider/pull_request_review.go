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
