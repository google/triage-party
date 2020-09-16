package provider

import (
	"time"
)

// IssueComment represents a comment left on an issue.
type IssueComment struct {
	ID     *int64  `json:"id,omitempty"`
	NodeID *string `json:"node_id,omitempty"`
	Body   *string `json:"body,omitempty"`
	User   *User   `json:"user,omitempty"`

	Reactions *Reactions `json:"reactions,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// AuthorAssociation is the comment author's relationship to the issue's repository.
	// Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".
	AuthorAssociation *string `json:"author_association,omitempty"`
	URL               *string `json:"url,omitempty"`
	HTMLURL           *string `json:"html_url,omitempty"`
	IssueURL          *string `json:"issue_url,omitempty"`
}

// GetAuthorAssociation returns the AuthorAssociation field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetAuthorAssociation() string {
	if i == nil || i.AuthorAssociation == nil {
		return ""
	}
	return *i.AuthorAssociation
}

// GetBody returns the Body field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetBody() string {
	if i == nil || i.Body == nil {
		return ""
	}
	return *i.Body
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetCreatedAt() time.Time {
	if i == nil || i.CreatedAt == nil {
		return time.Time{}
	}
	return *i.CreatedAt
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetHTMLURL() string {
	if i == nil || i.HTMLURL == nil {
		return ""
	}
	return *i.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetID() int64 {
	if i == nil || i.ID == nil {
		return 0
	}
	return *i.ID
}

// GetReactions returns the Reactions field.
func (i *IssueComment) GetReactions() *Reactions {
	if i == nil {
		return nil
	}
	return i.Reactions
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetUpdatedAt() time.Time {
	if i == nil || i.UpdatedAt == nil {
		return time.Time{}
	}
	return *i.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (i *IssueComment) GetURL() string {
	if i == nil || i.URL == nil {
		return ""
	}
	return *i.URL
}

// GetUser returns the User field.
func (i *IssueComment) GetUser() *User {
	if i == nil {
		return nil
	}
	return i.User
}

func (i IssueComment) String() string {
	return Stringify(i)
}
