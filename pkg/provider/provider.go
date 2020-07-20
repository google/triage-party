package provider

import (
	"context"
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/triage"
	"golang.org/x/oauth2"
)

const (
	_ = iota
	GithubProviderType
	GitlabProviderType
)

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
)

type Config struct {
	GithubAPIRawURL *string
	GithubTokenFile *string
}

func InitClients(ctx context.Context, c Config) {
	initGithubClient(ctx, c)
	initGitlabClient(ctx, c)
}
