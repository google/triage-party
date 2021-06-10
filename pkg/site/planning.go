// Copyright 2021 Google Inc.
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

package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/provider"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

const (
	unplanned  = "Not Planned"
	inProgress = "In Progress"
)

// Planning shows a view of a collection.
func (h *Handlers) Planning() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":          toJS,
		"toYAML":        toYAML,
		"toJSfunc":      toJSfunc,
		"toDays":        toDays,
		"HumanDuration": humanDuration,
		"UnixNano":      unixNano,
		"getAssignees":  getAssignees,
		"getMilestone":  getMilestone,
		"Avatar":        avatarSticky,
		"Class":         className,
	}

	t := template.Must(template.New("planning").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "planning.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("GET %s: %v", r.URL.Path, r.Header)

		id := strings.TrimPrefix(r.URL.Path, "/p/")

		p, err := h.collectionPage(r.Context(), id, isRefresh(r))
		if err != nil {
			http.Error(w, fmt.Sprintf("planning page for %q: %v", id, err), 500)
			klog.Errorf("page: %v", err)

			return
		}

		if p.CollectionResult.RuleResults != nil {
			p.Description = p.Collection.Description
			p.Swimlanes = groupByState(p.CollectionResult.RuleResults)
		}

		err = t.ExecuteTemplate(w, "base", p)

		if err != nil {
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}

func getMilestone(c *hubbub.Conversation) string {
	if c.Milestone == nil || len(c.PullRequestRefs) == 0 {
		return ""
	}
	return fmt.Sprintf("Milestone: %s", *c.Milestone.Title)
}

func avatarSticky(u *provider.User) template.HTML {
	if u.GetLogin() == unassigned {
		return `<div class="sticky-reaction">🤷</div>`
	}
	return avatar(u)
}

func getAssignees(co *hubbub.Conversation) []*provider.User {
	if co.Assignees == nil || len(co.Assignees) == 0 {
		return []*provider.User{{Login: &unassigned}}
	}
	return co.Assignees
}

func groupByState(results []*triage.RuleResult) []*Swimlane {
	lanes := map[string]*Swimlane{
		unplanned: {
			Name:    unplanned,
			Columns: make([]*triage.RuleResult, len(results)),
		},
		inProgress: {
			Name:    inProgress,
			Columns: make([]*triage.RuleResult, len(results)),
		},
	}
	seen := map[int]struct{}{}
	for i, r := range results {
		for _, co := range r.Items {
			if _, ok := seen[co.ID]; ok {
				continue
			}
			seen[co.ID] = struct{}{}
			var state string
			if co.Milestone == nil {
				state = unplanned
			} else {
				if len(co.PullRequestRefs) > 0 {
					state = inProgress
				} else {
					state = *co.Milestone.Title
				}
			}
			if lanes[state] == nil {
				lanes[state] = &Swimlane{
					Name:    state,
					Description: fmt.Sprintf("Due on %s (%d/%d) open",
						co.Milestone.GetDueOn().Format("2020-01-02"), co.Milestone.GetOpenIssues(),
						co.Milestone.GetOpenIssues() + co.Milestone.GetClosedIssues()),
					Url:     *co.Milestone.HTMLURL,
					Columns: make([]*triage.RuleResult, len(results)),
				}
			}
			if lanes[state].Columns[i] == nil {
				lanes[state].Columns[i] = &triage.RuleResult{
					Rule:  r.Rule,
					Items: []*hubbub.Conversation{},
				}
			}
			lanes[state].Columns[i].Items = append(lanes[state].Columns[i].Items, co)
			lanes[state].Issues++
		}
	}

	sl := []*Swimlane{lanes[unplanned]}

	for k, v := range lanes {
		if k == unplanned || k == inProgress {
			continue
		}
		sl = append(sl, v)
	}
	return append(sl, lanes[inProgress])
}
