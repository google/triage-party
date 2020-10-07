// Copyright 2020 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"fmt"
	"github.com/google/triage-party/pkg/constants"
	"io/ioutil"
	"k8s.io/klog/v2"
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
	if (githubProvider == nil) && (gitlabProvider == nil) {
		klog.Exitf("You should use at least 1 provider: gitlab/github")
	}
}

func ResolveProviderByHost(providerHost string) Provider {
	switch providerHost {
	case constants.GithubProviderHost:
		if githubProvider == nil {
			klog.Exitf("You need initialize github provider")
		}
		return githubProvider
	case constants.GitlabProviderHost:
		if gitlabProvider == nil {
			klog.Exitf("You need initialize gitlab provider")
		}
		return gitlabProvider
	}
	fmt.Println("not existing provider")
	return nil
}

func mustReadToken(path string, token, env, providerName string) string {
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = string(t)
		klog.Infof("loaded %d byte %s token from %s", len(token), providerName, path)
	} else {
		klog.Infof("loaded %d byte %s token from %s", len(token), providerName, env)
	}

	token = strings.TrimSpace(token)
	if len(token) < 8 {
		klog.Exitf("%s token impossibly small: %q", providerName, token)
	}
	return token
}
