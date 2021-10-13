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
	unplanned  = "Unplanned"
)

// Planning shows a view of a collection.
// Planning board can be used planning Sprints and futures releases
// categorized in swimlanes defining a feature or Objective.
// The planning board gives a view of all planned milestones with issues
// versus the current or selected milestone in the Kanban Board.
// It also highlights all new incoming issues not assigned to a milestone.

func (h *Handlers) Planning() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":               toJS,
		"toYAML":             toYAML,
		"toJSfunc":           toJSfunc,
		"toDays":             toDays,
		"HumanDuration":      humanDuration,
		"UnixNano":           unixNano,
		"getAssignees":       getAssignees,
		"unAssignedOrAvatar": unAssignedOrAvatar,
		"Class":              className,
		"isUnplanned":        isUnplanned,
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

// unAssignedOrAvatar is used in a sticky note and hence
// wrapping "unAssigned
func unAssignedOrAvatar(u *provider.User) template.HTML {
	if u.GetLogin() == unassigned {
		return `🤷`
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
					state = *co.Milestone.Title
			}
			if lanes[state] == nil {
				ts := co.Milestone.GetDueOn()
				lanes[state] = &Swimlane{
					Name: state,
					Description: fmt.Sprintf("Due on %s-%d (%d/%d) open",
						ts.Month().String(), ts.Day(), co.Milestone.GetOpenIssues(),
						co.Milestone.GetOpenIssues()+co.Milestone.GetClosedIssues()),
					URL:     *co.Milestone.HTMLURL,
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
		if k == unplanned {
			continue
		}
		sl = append(sl, v)
	}
	return sl
}

func isUnplanned(name string) bool {
	return name == unplanned
}
