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
	"testing"

	"github.com/google/go-github/v33/github"
	"github.com/stretchr/testify/assert"
)

func TestGitHub_GetResponse(t *testing.T) {
	p := GitHubProvider{}
	p.getResponse(nil)
}

func TestGitHub_GetIssues(t *testing.T) {
	p := GitHubProvider{}
	p.getIssues(nil)
}

func TestGitHub_GetIssueComments(t *testing.T) {
	p := GitHubProvider{}
	p.getIssueComments(nil)
}

func TestGitHub_GetIssueTimeline(t *testing.T) {
	p := GitHubProvider{}
	p.getIssueTimeline(nil)
}

func TestGitHub_GetPullRequestsList(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestsList(nil)
}

func TestGitHub_GetPullRequest(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequest(nil)
}

func TestGitHub_GetPullRequestListComments(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestListComments(nil)
}

func TestGitHub_GetPullRequestsListReviews(t *testing.T) {
	p := GitHubProvider{}
	p.getPullRequestsListReviews(nil)
}

func TestGitHub_GetIssueListByRepoOptions(t *testing.T) {
	p := GitHubProvider{}

	tests := []struct {
		name string
		sp   SearchParams
		want *github.IssueListByRepoOptions
	}{
		{
			name: "no labels",
			sp:   SearchParams{},
			want: &github.IssueListByRepoOptions{},
		},
		{
			name: "with labels",
			sp: SearchParams{
				Repo: Repo{
					Labels: []string{"bug", "p0"},
				},
			},
			want: &github.IssueListByRepoOptions{
				Labels: []string{"bug", "p0"},
			},
		},
		{
			name: "with other options and labels",
			sp: SearchParams{
				Repo: Repo{
					Labels: []string{"feature"},
				},
				IssueListByRepoOptions: IssueListByRepoOptions{
					State: "closed",
				},
			},
			want: &github.IssueListByRepoOptions{
				State:  "closed",
				Labels: []string{"feature"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.getIssueListByRepoOptions(tt.sp)
			assert.Equal(t, tt.want, got)
		})
	}
}
