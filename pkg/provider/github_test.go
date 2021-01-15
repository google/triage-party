package provider

import (
	"testing"
)

func TestGitHub_GetResponse(t *testing.T) {
	p := GitHubProvider{}
	p.getResponse(nil)
}

func TestGitHub_GetIssues(t *testing.T) {
	p := GitHubProvider{}
	p.getIssues(nil)
}

func TestGitHub_GetIssueComments(t *testing.T) {
	p := GitHubProvider{}
	p.getIssueComments(nil)
}

func TestGitHub_GetIssueTimeline(t *testing.T) {
	p := GitHubProvider{}
	p.getIssueTimeline(nil)
}

func TestGitHub_GetPullRequestsList(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestsList(nil)
}

func TestGitHub_GetPullRequest(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequest(nil)
}

func TestGitHub_GetPullRequestListComments(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestListComments(nil)
}

func TestGitHub_GetPullRequestsListReviews(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestsListReviews(nil)
}
