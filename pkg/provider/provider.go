package provider

import (
	"context"
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/triage"
	"golang.org/x/oauth2"
)

// TODO do we need these constants?
const (
	_ = iota
	GithubProviderType
	GitlabProviderType
)

const (
	GithubProviderHost = "github.com"
	GitlabProviderHost = "gitlab.com"
)

type Provider interface {
	GetClient() interface{}
	IssuesListByRepo() ([]*models.Issue, *models.Response, error)
}

type Provider struct {
	DataProvider
	client interface{}
}

func initGithubClient(ctx context.Context, c Config) {
	githubClient = triage.MustCreateGithubClient(*c.GithubAPIRawURL, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: triage.MustReadToken(*c.GithubTokenFile, "GITHUB_TOKEN")},
	)))
}

func initGitlabClient(ctx context.Context, c Config) {
	// TODO
}

var (
	githubClient *github.Client
	gitlabClient interface{} //TODO
)

type Config struct {
	GithubAPIRawURL *string
	GithubTokenFile *string
}

func InitClients(ctx context.Context, c Config) {
	initGithubClient(ctx, c)
	initGitlabClient(ctx, c)
}

// TODO do we need this method?
func ResolveProviderByType(providerType int) *Provider {
	var client interface{}

	switch providerType {
	case GithubProviderType:
		client = githubClient
	case GitlabProviderType:
		client = gitlabClient
	}
	return &Provider{
		client: client,
	}
}

func ResolveProviderByHost(providerHost string) *Provider {
	var client interface{}

	switch providerHost {
	case GithubProviderHost:
		client = githubClient
	case GitlabProviderHost:
		client = gitlabClient
	}
	return &Provider{
		client: client,
	}
}
