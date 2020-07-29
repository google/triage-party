package provider

import "time"

type Thing struct {
	Created time.Time

	PullRequests        []*PullRequest
	Issues              []*Issue
	PullRequestComments []*PullRequestComment
	IssueComments       []*IssueComment
	Timeline            []*Timeline
	Reviews             []*PullRequestReview
	StringBool          map[string]bool
}
