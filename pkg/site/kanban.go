// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package handlers define HTTP handlers.
package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

type Swimlane struct {
	User    *github.User
	Columns []*triage.RuleResult
}

func groupByUser(results []*triage.RuleResult) []*Swimlane {
	lanes := map[string]*Swimlane{}

	for i, r := range results {
		for _, co := range r.Items {
			for _, a := range co.Assignees {
				assignee := a.GetLogin()
				if lanes[assignee] == nil {
					lanes[assignee] = &Swimlane{
						User:    a,
						Columns: make([]*triage.RuleResult, len(results)),
					}
				}

				if lanes[assignee].Columns[i] == nil {
					lanes[assignee].Columns[i] = &triage.RuleResult{
						Rule:  r.Rule,
						Items: []*hubbub.Conversation{},
					}
				}
				lanes[assignee].Columns[i].Items = append(lanes[assignee].Columns[i].Items, co)
			}
		}
	}

	ls := []*Swimlane{}
	for _, v := range lanes {
		ls = append(ls, v)
	}

	return ls
}

// Kanban shows a kanban swimlane view of a collection
func (h *Handlers) Kanban() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":          toJS,
		"toYAML":        toYAML,
		"toJSfunc":      toJSfunc,
		"toDays":        toDays,
		"HumanDuration": humanDuration,
		"HumanTime":     humanTime,
		"UnixNano":      unixNano,
		"Avatar":        avatarWide,
		"Class":         className,
	}

	t := template.Must(template.New("kanban").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "kanban.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/k/")

		p, err := h.collectionPage(r.Context(), id, isRefresh(r))
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("page: %v", err)
			return
		}

		p.Swimlanes = groupByUser(p.CollectionResult.RuleResults)

		klog.V(2).Infof("page context: %+v", p)
		err = t.ExecuteTemplate(w, "base", p)
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}
