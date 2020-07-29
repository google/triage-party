package provider

import (
	"context"
	"fmt"
	"github.com/google/triage-party/pkg/constants"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

type Provider interface {
	IssuesListByRepo(ctx context.Context, sp SearchParams) ([]*Issue, *Response, error)
	IssuesListComments(ctx context.Context, sp SearchParams) ([]*IssueComment, *Response, error)
	IssuesListIssueTimeline(ctx context.Context, sp SearchParams) ([]*Timeline, *Response, error)
	PullRequestsList(ctx context.Context, sp SearchParams) ([]*PullRequest, *Response, error)
	PullRequestsGet(ctx context.Context, sp SearchParams) (*PullRequest, *Response, error)
	PullRequestsListComments(ctx context.Context, sp SearchParams) ([]*PullRequestComment, *Response, error)
	PullRequestsListReviews(ctx context.Context, sp SearchParams) ([]*PullRequestReview, *Response, error)
}

var (
	githubProvider *GithubProvider
	gitlabProvider *GitlabProvider
)

type Config struct {
	GithubAPIRawURL *string
	GithubTokenFile *string
	GitlabTokenFile *string
}

func InitProviders(ctx context.Context, c Config) {
	initGithub(ctx, c)
	initGitlab(c)
}

func ResolveProviderByHost(providerHost string) Provider {
	switch providerHost {
	case constants.GithubProviderHost:
		return githubProvider
	case constants.GitlabProviderHost:
		return gitlabProvider
	}
	fmt.Println("not existing provider")
	return nil
}

func mustReadToken(path string, env string) string {
	token := os.Getenv(env)
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = string(t)
		klog.Infof("loaded %d byte github/gitlab token from %s", len(token), path)
	} else {
		klog.Infof("loaded %d byte github/gitlab token from %s", len(token), env)
	}

	token = strings.TrimSpace(token)
	if len(token) < 8 {
		klog.Exitf("github/gitlab token impossibly small: %q", token)
	}
	return token
}
