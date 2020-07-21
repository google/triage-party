package models

import "time"

// PullRequestReview represents a review of a pull request.
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

// GetAuthorAssociation returns the AuthorAssociation field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetAuthorAssociation() string {
	if p == nil || p.AuthorAssociation == nil {
		return ""
	}
	return *p.AuthorAssociation
}

// GetBody returns the Body field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetBody() string {
	if p == nil || p.Body == nil {
		return ""
	}
	return *p.Body
}

// GetCommitID returns the CommitID field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetCommitID() string {
	if p == nil || p.CommitID == nil {
		return ""
	}
	return *p.CommitID
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetHTMLURL() string {
	if p == nil || p.HTMLURL == nil {
		return ""
	}
	return *p.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetID() int64 {
	if p == nil || p.ID == nil {
		return 0
	}
	return *p.ID
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetNodeID() string {
	if p == nil || p.NodeID == nil {
		return ""
	}
	return *p.NodeID
}

// GetPullRequestURL returns the PullRequestURL field if it's non-nil, zero value otherwise.
func (p *PullRequestReview) GetPullRequestURL() string {
	if p == nil || p.PullRequestURL == nil {
		return ""
	}
	return *p.PullRequestURL
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

// GetUser returns the User field.
func (p *PullRequestReview) GetUser() *User {
	if p == nil {
		return nil
	}
	return p.User
}
