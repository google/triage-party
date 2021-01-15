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
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/klog/v2"
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

type Config struct {
	GitHubAPIURL    string
	GitHubTokenPath string

	GitLabTokenPath string
}

func ReadToken(path string, envVar string) string {
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token := strings.TrimSpace(string(t))
		klog.Infof("loaded %d byte %s token from %s", len(token), path)
		return token
	}

	token := strings.TrimSpace(os.Getenv(envVar))
	if token == "" {
		klog.Warningf("No token found in environment variable %s (empty)", envVar)
	} else {
		klog.Infof("loaded %d byte %s token from %s", len(token), envVar)
	}
	return token
}
