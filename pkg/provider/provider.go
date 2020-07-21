package provider

import (
	"context"
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/triage"
	"golang.org/x/oauth2"
)

const (
	GithubProviderHost = "github.com"
	GitlabProviderHost = "gitlab.com"
)

type Provider interface {
	IssuesListByRepo(sp models.SearchParams) ([]*models.Issue, *models.Response, error)
	IssuesListComments(sp models.SearchParams) ([]*models.IssueComment, *models.Response, error)
	IssuesListIssueTimeline(sp models.SearchParams) ([]*models.Timeline, *models.Response, error)
	PullRequestsList(sp models.SearchParams) ([]*models.PullRequest, *models.Response, error)
	PullRequestsGet(sp models.SearchParams) (*models.PullRequest, *models.Response, error)
	PullRequestsListComments(sp models.SearchParams) ([]*models.PullRequestComment, *models.Response, error)
	PullRequestsListReviews(sp models.SearchParams) ([]*models.PullRequestReview, *models.Response, error)
}

var (
	githubProvider *GithubProvider
)

func initGithub(ctx context.Context, c Config) {
	cl := triage.MustCreateGithubClient(*c.GithubAPIRawURL, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: triage.MustReadToken(*c.GithubTokenFile, "GITHUB_TOKEN")},
	)))
	githubProvider = &GithubProvider{
		client: cl,
	}
}

func initGitlab(ctx context.Context, c Config) {
	// TODO
}

type Config struct {
	GithubAPIRawURL *string
	GithubTokenFile *string
}

func InitProviders(ctx context.Context, c Config) {
	initGithub(ctx, c)
	initGitlab(ctx, c)
}

func ResolveProviderByHost(providerHost string) Provider {
	// TODO implement gitlab
	return githubProvider
	//switch providerHost {
	//case GithubProviderHost:
	//	return githubProvider
	//case GitlabProviderHost:
	//	return nil //TODO
	//}
}
