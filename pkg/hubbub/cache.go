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

package hubbub

import (
	"fmt"
	"github.com/google/triage-party/pkg/provider"
)

// issueSearchKey is the cache key used for issues
func issueSearchKey(sp provider.SearchParams) string {
	if sp.UpdateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-issues-within-%.1fh", sp.Repo.Organization, sp.Repo.Project, sp.State, sp.UpdateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-issues", sp.Repo.Organization, sp.Repo.Project, sp.State)
}

// prSearchKey is the cache key used for prs
func prSearchKey(sp provider.SearchParams) string {
	if sp.UpdateAge > 0 {
		return fmt.Sprintf("%s-%s-%s-prs-within-%.1fh", sp.Repo.Organization, sp.Repo.Project, sp.State, sp.UpdateAge.Hours())
	}
	return fmt.Sprintf("%s-%s-%s-prs", sp.Repo.Organization, sp.Repo.Project, sp.State)
}
