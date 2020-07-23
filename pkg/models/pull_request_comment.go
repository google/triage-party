package models

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

// GetCommitID returns the CommitID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetCommitID() string {
	if p == nil || p.CommitID == nil {
		return ""
	}
	return *p.CommitID
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetCreatedAt() time.Time {
	if p == nil || p.CreatedAt == nil {
		return time.Time{}
	}
	return *p.CreatedAt
}

// GetDiffHunk returns the DiffHunk field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetDiffHunk() string {
	if p == nil || p.DiffHunk == nil {
		return ""
	}
	return *p.DiffHunk
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

// GetInReplyTo returns the InReplyTo field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetInReplyTo() int64 {
	if p == nil || p.InReplyTo == nil {
		return 0
	}
	return *p.InReplyTo
}

// GetLine returns the Line field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetLine() int {
	if p == nil || p.Line == nil {
		return 0
	}
	return *p.Line
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetNodeID() string {
	if p == nil || p.NodeID == nil {
		return ""
	}
	return *p.NodeID
}

// GetOriginalCommitID returns the OriginalCommitID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetOriginalCommitID() string {
	if p == nil || p.OriginalCommitID == nil {
		return ""
	}
	return *p.OriginalCommitID
}

// GetOriginalLine returns the OriginalLine field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetOriginalLine() int {
	if p == nil || p.OriginalLine == nil {
		return 0
	}
	return *p.OriginalLine
}

// GetOriginalPosition returns the OriginalPosition field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetOriginalPosition() int {
	if p == nil || p.OriginalPosition == nil {
		return 0
	}
	return *p.OriginalPosition
}

// GetOriginalStartLine returns the OriginalStartLine field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetOriginalStartLine() int {
	if p == nil || p.OriginalStartLine == nil {
		return 0
	}
	return *p.OriginalStartLine
}

// GetPath returns the Path field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetPath() string {
	if p == nil || p.Path == nil {
		return ""
	}
	return *p.Path
}

// GetPosition returns the Position field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetPosition() int {
	if p == nil || p.Position == nil {
		return 0
	}
	return *p.Position
}

// GetPullRequestReviewID returns the PullRequestReviewID field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetPullRequestReviewID() int64 {
	if p == nil || p.PullRequestReviewID == nil {
		return 0
	}
	return *p.PullRequestReviewID
}

// GetPullRequestURL returns the PullRequestURL field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetPullRequestURL() string {
	if p == nil || p.PullRequestURL == nil {
		return ""
	}
	return *p.PullRequestURL
}

// GetReactions returns the Reactions field.
func (p *PullRequestComment) GetReactions() *Reactions {
	if p == nil {
		return nil
	}
	return p.Reactions
}

// GetSide returns the Side field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetSide() string {
	if p == nil || p.Side == nil {
		return ""
	}
	return *p.Side
}

// GetStartLine returns the StartLine field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetStartLine() int {
	if p == nil || p.StartLine == nil {
		return 0
	}
	return *p.StartLine
}

// GetStartSide returns the StartSide field if it's non-nil, zero value otherwise.
func (p *PullRequestComment) GetStartSide() string {
	if p == nil || p.StartSide == nil {
		return ""
	}
	return *p.StartSide
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
