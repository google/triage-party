package provider

import (
	"testing"
)

func TestGitLab_GetResponse(t *testing.T) {
	p := GitLabProvider{}
	p.getResponse(nil)
}

func TestGitLab_GetIssues(t *testing.T) {
	p := GitLabProvider{}
	p.getIssues(nil)
}

func TestGitLab_GetIssueComments(t *testing.T) {
	p := GitLabProvider{}
	p.getIssueComments(nil)
}

func TestGitLab_GetPullRequests(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequests(nil)
}

func TestGitLab_GetPullRequest(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequest(nil)
}

func TestGitLab_GetPullRequestComments(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequestComments(nil)
}

func TestGitLab_GetPullRequestReviews(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequestReviews(nil)
}
