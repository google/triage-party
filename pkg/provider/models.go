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

import (
	"time"
)

type Response struct {
	// github specific
	// These fields provide the page values for paginating through a set of
	// results. Any or all of these may be set to the zero value for
	// responses that are not part of a paginated set, or for which there
	// are no additional pages.
	//
	// These fields support what is called "offset pagination" and should
	// be used with the ListOptions struct.
	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int

	// Additionally, some APIs support "cursor pagination" instead of offset.
	// This means that a token points directly to the next record which
	// can lead to O(1) performance compared to O(n) performance provided
	// by offset pagination.
	//
	// For APIs that support cursor pagination (such as
	// TeamsService.ListIDPGroupsInOrganization), the following field
	// will be populated to point to the next page.
	//
	// To use this token, set ListCursorOptions.Page to this value before
	// calling the endpoint again.
	NextPageToken string

	// Explicitly specify the Rate type so Rate's String() receiver doesn't
	// propagate to Response.
	Rate Rate
}

type Repo struct {
	Organization string
	Project      string
	Host         string
	Group        string
}

type SearchParams struct {
	Repo        Repo
	Filters     []Filter
	NewerThan   time.Time
	Hidden      bool
	State       string
	UpdateAge   time.Duration
	UpdateAt    time.Time
	Age         time.Time
	SearchKey   string
	IssueNumber int
	Fetch       bool

	IssueListByRepoOptions   IssueListByRepoOptions
	IssueListCommentsOptions IssueListCommentsOptions
	ListOptions              ListOptions
	PullRequestListOptions   PullRequestListOptions
}

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

// abstraction model for github.Rate struct
// Rate represents the rate limit for the current client.
type Rate struct {
	// The number of requests per hour the client is currently limited to.
	Limit int `json:"limit"`

	// The number of remaining requests the client can make this hour.
	Remaining int `json:"remaining"`

	// The time at which the current rate limit will reset.
	Reset Timestamp `json:"reset"`
}

// abstraction model for github.IssueListCommentsOptions struct
// IssueListCommentsOptions specifies the optional parameters to the
// IssuesService.ListComments method.
type IssueListCommentsOptions struct {
	// Sort specifies how to sort comments. Possible values are: created, updated.
	Sort *string `url:"sort,omitempty"`

	// Direction in which to sort comments. Possible values are: asc, desc.
	Direction *string `url:"direction,omitempty"`

	// Since filters comments by time.
	Since *time.Time `url:"since,omitempty"`

	ListOptions
}

// PullRequestListOptions specifies the optional parameters to the
// PullRequestsService.List method.
type PullRequestListOptions struct {
	// State filters pull requests based on their state. Possible values are:
	// open, closed, all. Default is "open".
	State string `url:"state,omitempty"`

	// Head filters pull requests by head user and branch name in the format of:
	// "user:ref-name".
	Head string `url:"head,omitempty"`

	// Base filters pull requests by base branch name.
	Base string `url:"base,omitempty"`

	// Sort specifies how to sort pull requests. Possible values are: created,
	// updated, popularity, long-running. Default is "created".
	Sort string `url:"sort,omitempty"`

	// Direction in which to sort pull requests. Possible values are: asc, desc.
	// If Sort is "created" or not specified, Default is "desc", otherwise Default
	// is "asc"
	Direction string `url:"direction,omitempty"`

	ListOptions
}

// PullRequestListCommentsOptions specifies the optional parameters to the
// PullRequestsService.ListComments method.
type PullRequestListCommentsOptions struct {
	// Sort specifies how to sort comments. Possible values are: created, updated.
	Sort string `url:"sort,omitempty"`

	// Direction in which to sort comments. Possible values are: asc, desc.
	Direction string `url:"direction,omitempty"`

	// Since filters comments by time.
	Since time.Time `url:"since,omitempty"`

	ListOptions
}
