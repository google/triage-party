package models

import (
	"github.com/google/triage-party/pkg/utils"
	"time"
)

// Issue represents a GitHub issue on a repository.
//
// Note: As far as the GitHub API is concerned, every pull request is an issue,
// but not every issue is a pull request. Some endpoints, events, and webhooks
// may also return pull requests via this struct. If PullRequestLinks is nil,
// this is an issue, and if PullRequestLinks is not nil, this is a pull request.
// The IsPullRequest helper method can be used to check that.
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

	// TextMatches is only populated from search results that request text matches
	// See: search.go and https://developer.github.com/v3/search/#text-match-metadata
	//TextMatches []*TextMatch `json:"text_matches,omitempty"`

	// ActiveLockReason is populated only when LockReason is provided while locking the issue.
	// Possible values are: "off-topic", "too heated", "resolved", and "spam".
	ActiveLockReason *string `json:"active_lock_reason,omitempty"`
}

// GetActiveLockReason returns the ActiveLockReason field if it's non-nil, zero value otherwise.
func (i *Issue) GetActiveLockReason() string {
	if i == nil || i.ActiveLockReason == nil {
		return ""
	}
	return *i.ActiveLockReason
}

// GetAssignee returns the Assignee field.
func (i *Issue) GetAssignee() *User {
	if i == nil {
		return nil
	}
	return i.Assignee
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

// GetCommentsURL returns the CommentsURL field if it's non-nil, zero value otherwise.
func (i *Issue) GetCommentsURL() string {
	if i == nil || i.CommentsURL == nil {
		return ""
	}
	return *i.CommentsURL
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (i *Issue) GetCreatedAt() time.Time {
	if i == nil || i.CreatedAt == nil {
		return time.Time{}
	}
	return *i.CreatedAt
}

// GetEventsURL returns the EventsURL field if it's non-nil, zero value otherwise.
func (i *Issue) GetEventsURL() string {
	if i == nil || i.EventsURL == nil {
		return ""
	}
	return *i.EventsURL
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

// GetLabelsURL returns the LabelsURL field if it's non-nil, zero value otherwise.
func (i *Issue) GetLabelsURL() string {
	if i == nil || i.LabelsURL == nil {
		return ""
	}
	return *i.LabelsURL
}

// GetLocked returns the Locked field if it's non-nil, zero value otherwise.
func (i *Issue) GetLocked() bool {
	if i == nil || i.Locked == nil {
		return false
	}
	return *i.Locked
}

// GetMilestone returns the Milestone field.
func (i *Issue) GetMilestone() *Milestone {
	if i == nil {
		return nil
	}
	return i.Milestone
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (i *Issue) GetNodeID() string {
	if i == nil || i.NodeID == nil {
		return ""
	}
	return *i.NodeID
}

// GetNumber returns the Number field if it's non-nil, zero value otherwise.
func (i *Issue) GetNumber() int {
	if i == nil || i.Number == nil {
		return 0
	}
	return *i.Number
}

// GetPullRequestLinks returns the PullRequestLinks field.
func (i *Issue) GetPullRequestLinks() *PullRequestLinks {
	if i == nil {
		return nil
	}
	return i.PullRequestLinks
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

// GetRepositoryURL returns the RepositoryURL field if it's non-nil, zero value otherwise.
func (i *Issue) GetRepositoryURL() string {
	if i == nil || i.RepositoryURL == nil {
		return ""
	}
	return *i.RepositoryURL
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
	return utils.Stringify(i)
}

// IsPullRequest reports whether the issue is also a pull request. It uses the
// method recommended by GitHub's API documentation, which is to check whether
// PullRequestLinks is non-nil.
func (i Issue) IsPullRequest() bool {
	return i.PullRequestLinks != nil
}
