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

// PullRequestLinks object is added to the Issue object when it's an issue included
// in the IssueCommentEvent webhook payload, if the webhook is fired by a comment on a PR.
type PullRequestLinks struct {
	URL      *string `json:"url,omitempty"`
	HTMLURL  *string `json:"html_url,omitempty"`
	DiffURL  *string `json:"diff_url,omitempty"`
	PatchURL *string `json:"patch_url,omitempty"`
}
