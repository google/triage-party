package provider

import (
	"testing"
)

func TestGitlab_GetResponse(t *testing.T) {
	p := GitlabProvider{}
	p.getResponse(nil)
}

func TestGitlab_GetIssues(t *testing.T) {
	p := GitlabProvider{}
	p.getIssues(nil)
}

func TestGitlab_GetIssueComments(t *testing.T) {
	p := GitlabProvider{}
	p.getIssueComments(nil)
}

func TestGitlab_GetPullRequests(t *testing.T) {
	p := GitlabProvider{}
	p.getPullRequests(nil)
}

func TestGitlab_GetPullRequest(t *testing.T) {
	p := GitlabProvider{}
	p.getPullRequest(nil)
}

func TestGitlab_GetPullRequestComments(t *testing.T) {
	p := GitlabProvider{}
	p.getPullRequestComments(nil)
}

func TestGitlab_GetPullRequestReviews(t *testing.T) {
	p := GitlabProvider{}
	p.getPullRequestReviews(nil)
}
