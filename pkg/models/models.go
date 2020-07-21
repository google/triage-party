package models

import (
	"context"
	"github.com/google/triage-party/pkg/hubbub"
	"time"
)

type Response struct {
}

type Repo struct {
	Organization string
	Project      string
	Host         string
}

type SearchParams struct {
	Repo        Repo
	Filters     []hubbub.Filter
	Ctx         context.Context
	NewerThan   time.Time
	Hidden      bool
	State       string
	UpdateAge   time.Duration
	SearchKey   string
	IssueNumber int
	Fetch       bool

	IssueListByRepoOptions IssueListByRepoOptions
}

// abstraction model for github.IssueListByRepoOptions struct
// IssueListByRepoOptions specifies the optional parameters to the
// IssuesService.ListByRepo method.
type IssueListByRepoOptions struct {
	// Milestone limits issues for the specified milestone. Possible values are
	// a milestone number, "none" for issues with no milestone, "*" for issues
	// with any milestone.
	Milestone string `url:"milestone,omitempty"`

	// State filters issues based on their state. Possible values are: open,
	// closed, all. Default is "open".
	State string `url:"state,omitempty"`

	// Assignee filters issues based on their assignee. Possible values are a
	// user name, "none" for issues that are not assigned, "*" for issues with
	// any assigned user.
	Assignee string `url:"assignee,omitempty"`

	// Creator filters issues based on their creator.
	Creator string `url:"creator,omitempty"`

	// Mentioned filters issues to those mentioned a specific user.
	Mentioned string `url:"mentioned,omitempty"`

	// Labels filters issues based on their label.
	Labels []string `url:"labels,omitempty,comma"`

	// Sort specifies how to sort issues. Possible values are: created, updated,
	// and comments. Default value is "created".
	Sort string `url:"sort,omitempty"`

	// Direction in which to sort issues. Possible values are: asc, desc.
	// Default is "desc".
	Direction string `url:"direction,omitempty"`

	// Since filters issues by time.
	Since time.Time `url:"since,omitempty"`

	ListOptions
}

// abstraction model for github.ListOptions struct
// ListOptions specifies the optional parameters to various List methods that
// support offset pagination.
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`

	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}
