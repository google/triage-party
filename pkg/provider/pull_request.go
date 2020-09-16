package provider

import (
	"time"
)

// PullRequest represents a GitHub pull request on a repository.
type PullRequest struct {
	ID                  *int64     `json:"id,omitempty"`
	Number              *int       `json:"number,omitempty"`
	State               *string    `json:"state,omitempty"`
	Locked              *bool      `json:"locked,omitempty"`
	Title               *string    `json:"title,omitempty"`
	Body                *string    `json:"body,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	MergedAt            *time.Time `json:"merged_at,omitempty"`
	Labels              []*Label   `json:"labels,omitempty"`
	User                *User      `json:"user,omitempty"`
	Draft               *bool      `json:"draft,omitempty"`
	Merged              *bool      `json:"merged,omitempty"`
	Mergeable           *bool      `json:"mergeable,omitempty"`
	MergeableState      *string    `json:"mergeable_state,omitempty"`
	MergedBy            *User      `json:"merged_by,omitempty"`
	MergeCommitSHA      *string    `json:"merge_commit_sha,omitempty"`
	Rebaseable          *bool      `json:"rebaseable,omitempty"`
	Comments            *int       `json:"comments,omitempty"`
	Commits             *int       `json:"commits,omitempty"`
	Additions           *int       `json:"additions,omitempty"`
	Deletions           *int       `json:"deletions,omitempty"`
	ChangedFiles        *int       `json:"changed_files,omitempty"`
	URL                 *string    `json:"url,omitempty"`
	HTMLURL             *string    `json:"html_url,omitempty"`
	IssueURL            *string    `json:"issue_url,omitempty"`
	StatusesURL         *string    `json:"statuses_url,omitempty"`
	DiffURL             *string    `json:"diff_url,omitempty"`
	PatchURL            *string    `json:"patch_url,omitempty"`
	CommitsURL          *string    `json:"commits_url,omitempty"`
	CommentsURL         *string    `json:"comments_url,omitempty"`
	ReviewCommentsURL   *string    `json:"review_comments_url,omitempty"`
	ReviewCommentURL    *string    `json:"review_comment_url,omitempty"`
	ReviewComments      *int       `json:"review_comments,omitempty"`
	Assignee            *User      `json:"assignee,omitempty"`
	Assignees           []*User    `json:"assignees,omitempty"`
	Milestone           *Milestone `json:"milestone,omitempty"`
	MaintainerCanModify *bool      `json:"maintainer_can_modify,omitempty"`
	AuthorAssociation   *string    `json:"author_association,omitempty"`
	NodeID              *string    `json:"node_id,omitempty"`
	RequestedReviewers  []*User    `json:"requested_reviewers,omitempty"`

	// RequestedTeams is populated as part of the PullRequestEvent.
	// See, https://developer.github.com/v3/activity/events/types/#pullrequestevent for an example.
	//RequestedTeams []*Team `json:"requested_teams,omitempty"`
	//
	//Links *PRLinks           `json:"_links,omitempty"`
	//Head  *PullRequestBranch `json:"head,omitempty"`
	//Base  *PullRequestBranch `json:"base,omitempty"`

	// ActiveLockReason is populated only when LockReason is provided while locking the pull request.
	// Possible values are: "off-topic", "too heated", "resolved", and "spam".
	ActiveLockReason *string `json:"active_lock_reason,omitempty"`
}

// GetAssignee returns the Assignee field.
func (p *PullRequest) GetAssignee() *User {
	if p == nil {
		return nil
	}
	return p.Assignee
}

// GetAuthorAssociation returns the AuthorAssociation field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetAuthorAssociation() string {
	if p == nil || p.AuthorAssociation == nil {
		return ""
	}
	return *p.AuthorAssociation
}

// GetBase returns the Base field.
//func (p *PullRequest) GetBase() *PullRequestBranch {
//	if p == nil {
//		return nil
//	}
//	return p.Base
//}

// GetBody returns the Body field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetBody() string {
	if p == nil || p.Body == nil {
		return ""
	}
	return *p.Body
}

// GetClosedAt returns the ClosedAt field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetClosedAt() time.Time {
	if p == nil || p.ClosedAt == nil {
		return time.Time{}
	}
	return *p.ClosedAt
}

// GetComments returns the Comments field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetComments() int {
	if p == nil || p.Comments == nil {
		return 0
	}
	return *p.Comments
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetCreatedAt() time.Time {
	if p == nil || p.CreatedAt == nil {
		return time.Time{}
	}
	return *p.CreatedAt
}

// GetDraft returns the Draft field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetDraft() bool {
	if p == nil || p.Draft == nil {
		return false
	}
	return *p.Draft
}

// GetHead returns the Head field.
//func (p *PullRequest) GetHead() *PullRequestBranch {
//	if p == nil {
//		return nil
//	}
//	return p.Head
//}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetHTMLURL() string {
	if p == nil || p.HTMLURL == nil {
		return ""
	}
	return *p.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetID() int64 {
	if p == nil || p.ID == nil {
		return 0
	}
	return *p.ID
}

// GetMerged returns the Merged field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetMerged() bool {
	if p == nil || p.Merged == nil {
		return false
	}
	return *p.Merged
}

// GetMergedBy returns the MergedBy field.
func (p *PullRequest) GetMergedBy() *User {
	if p == nil {
		return nil
	}
	return p.MergedBy
}

// GetMilestone returns the Milestone field.
func (p *PullRequest) GetMilestone() *Milestone {
	if p == nil {
		return nil
	}
	return p.Milestone
}

// GetNumber returns the Number field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetNumber() int {
	if p == nil || p.Number == nil {
		return 0
	}
	return *p.Number
}

// GetState returns the State field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetState() string {
	if p == nil || p.State == nil {
		return ""
	}
	return *p.State
}

// GetTitle returns the Title field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetTitle() string {
	if p == nil || p.Title == nil {
		return ""
	}
	return *p.Title
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetUpdatedAt() time.Time {
	if p == nil || p.UpdatedAt == nil {
		return time.Time{}
	}
	return *p.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (p *PullRequest) GetURL() string {
	if p == nil || p.URL == nil {
		return ""
	}
	return *p.URL
}

// GetUser returns the User field.
func (p *PullRequest) GetUser() *User {
	if p == nil {
		return nil
	}
	return p.User
}

func (p PullRequest) String() string {
	return Stringify(p)
}
