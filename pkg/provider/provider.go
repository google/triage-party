package provider

import (
	"context"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/interfaces"
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/triage"
	"golang.org/x/oauth2"
)

const (
	GithubProviderHost = "github.com"
	GitlabProviderHost = "gitlab.com"
)

type Provider interface {
	IssuesListByRepo(sp models.SearchParams) ([]interfaces.IItem, *models.Response, error)
	IssuesListComments(sp models.SearchParams) ([]interfaces.IIssueComment, *models.Response, error)
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
	switch providerHost {
	case GithubProviderHost:
		return githubProvider
	case GitlabProviderHost:
		return nil //TODO
	}
}
