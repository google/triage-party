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
)

func TestGitLab_GetResponse(t *testing.T) {
	p := GitLabProvider{}
	p.getResponse(nil)
}

func TestGitLab_GetIssues(t *testing.T) {
	p := GitLabProvider{}
	p.getIssues(nil)
}

func TestGitLab_GetIssueComments(t *testing.T) {
	p := GitLabProvider{}
	p.getIssueComments(nil)
}

func TestGitLab_GetPullRequests(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequests(nil)
}

func TestGitLab_GetPullRequest(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequest(nil)
}

func TestGitLab_GetPullRequestComments(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequestComments(nil)
}

func TestGitLab_GetPullRequestReviews(t *testing.T) {
	p := GitLabProvider{}
	p.getPullRequestReviews(nil)
}
