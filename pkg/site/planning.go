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

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)



const(
	priority = "priority/"
)

// Planning shows a view of a collection.
func (h *Handlers) Planning() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":          toJS,
		"toYAML":        toYAML,
		"toJSfunc":      toJSfunc,
		"toDays":        toDays,
		"HumanDuration": humanDuration,
		"RoughTime":     roughTime,
		"LateTime":      lateTime,
		"UnixNano":      unixNano,
		"Avatar":        avatarWide,
		"Class":         className,
		"getPriority":   getPriority,
		"isPriorityLabel": isPriorityLabel,
		"TextColor":     textColor,
		"shdDisplayLabel": shdDisplayLabel,
		"labelMatchesRule" : labelMatchesRule,
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
			p.Swimlanes = groupByStatus(p.CollectionResult.RuleResults)
		}

		err = t.ExecuteTemplate(w, "base", p)

		if err != nil {
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}

func getPriority(l string) string {
	return strings.TrimPrefix(l, priority)
}


func isPriorityLabel(l string) bool {
	return strings.HasPrefix(l, priority)
}

func shdDisplayLabel(l string, rule triage.Rule) bool {
	if isPriorityLabel(l) {
		return false
	}
	return !labelMatchesRule(l, rule)
}

func labelMatchesRule(l string, rule triage.Rule) bool {
	for _, r := range rule.Filters {
		if r.RawLabel == l {
			return true
		}
	}
	return false
}

func groupByStatus(results []*triage.RuleResult) []*Swimlane {
	lanes := map[string]*Swimlane{
		constants.OpenState: {
			Status:    constants.OpenState,
			Columns: make([]*triage.RuleResult, len(results)),
		},
		constants.ClosedState: {
			Status:    constants.ClosedState,
			Columns: make([]*triage.RuleResult, len(results)),
		},
		"In Progress": {
			Status:    "In Progress",
			Columns: make([]*triage.RuleResult, len(results)),
		},
	}

	for i, r := range results {
		for _, co := range r.Items {
			var state string
			if len(co.PullRequestRefs) > 0  {
				state = "In Progress"
			} else if co.State == constants.ClosedState {
				state = constants.ClosedState
			} else {
				state = constants.OpenState
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

	return []*Swimlane{lanes[constants.OpenState],
		lanes["In Progress"],
		lanes[constants.ClosedState]}
}