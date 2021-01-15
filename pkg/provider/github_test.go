package provider

import (
	"testing"
)

func TestGithub_GetResponse(t *testing.T) {
	p := GithubProvider{}
	p.getResponse(nil)
}

func TestGithub_GetIssues(t *testing.T) {
	p := GithubProvider{}
	p.getIssues(nil)
}

func TestGithub_GetIssueComments(t *testing.T) {
	p := GithubProvider{}
	p.getIssueComments(nil)
}

func TestGithub_GetIssueTimeline(t *testing.T) {
	p := GithubProvider{}
	p.getIssueTimeline(nil)
}

func TestGithub_GetPullRequestsList(t *testing.T) {
	p := GithubProvider{}
	p.getPullRequestsList(nil)
}

func TestGithub_GetPullRequest(t *testing.T) {
	p := GithubProvider{}
	p.getPullRequest(nil)
}

func TestGithub_GetPullRequestListComments(t *testing.T) {
	p := GithubProvider{}
	p.getPullRequestListComments(nil)
}

func TestGithub_GetPullRequestsListReviews(t *testing.T) {
	p := GithubProvider{}
	p.getPullRequestsListReviews(nil)
}
