package provider

import (
	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/models"
	"github.com/xanzy/go-gitlab"
	"log"
)

type GitlabProvider struct {
	client *gitlab.Client
}

func initGitlab(c Config) {
	cl, err := gitlab.NewClient(mustReadToken(*c.GithubTokenFile, constants.GitlabTokenEnvVar))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	gitlabProvider = &GitlabProvider{
		client: cl,
	}
}

func (p *GitlabProvider) IssuesListByRepo(sp models.SearchParams) ([]*models.Issue, *models.Response, error) {

}

func (p *GitlabProvider) IssuesListComments(sp models.SearchParams) ([]*models.IssueComment, *models.Response, error) {

}

func (p *GitlabProvider) IssuesListIssueTimeline(sp models.SearchParams) ([]*models.Timeline, *models.Response, error) {

}

func (p *GitlabProvider) PullRequestsList(sp models.SearchParams) ([]*models.PullRequest, *models.Response, error) {

}

func (p *GitlabProvider) PullRequestsGet(sp models.SearchParams) (*models.PullRequest, *models.Response, error) {

}

func (p *GitlabProvider) PullRequestsListComments(sp models.SearchParams) ([]*models.PullRequestComment, *models.Response, error) {

}

func (p *GitlabProvider) PullRequestsListReviews(sp models.SearchParams) ([]*models.PullRequestReview, *models.Response, error) {

}
