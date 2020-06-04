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

package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

var unassigned = "zz_unassigned"

// Swimlane is a row in a Kanban display.
type Swimlane struct {
	User    *github.User
	Columns []*triage.RuleResult
}

func avatarWide(u *github.User) template.HTML {
	if u.GetLogin() == unassigned {
		return template.HTML(`<div class="unassigned" title="Unassigned work - free for the taking!"></div>`)
	}

	return template.HTML(fmt.Sprintf(`<a href="%s" title="%s"><img src="%s" width="96" height="96"></a>`, u.GetHTMLURL(), u.GetLogin(), u.GetAvatarURL()))
}

func groupByUser(results []*triage.RuleResult, milestoneID int) []*Swimlane {
	lanes := map[string]*Swimlane{}

	for i, r := range results {
		for _, co := range r.Items {
			assignees := co.Assignees
			if len(assignees) == 0 {
				assignees = append(assignees, &github.User{
					Login: &unassigned,
				})
			}

			for _, a := range assignees {
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

				if milestoneID == 0 || co.Milestone.GetNumber() == milestoneID {
					lanes[assignee].Columns[i].Items = append(lanes[assignee].Columns[i].Items, co)
				}
			}
		}
	}

	ls := []*Swimlane{}
	for _, v := range lanes {
		ls = append(ls, v)
	}

	return ls
}

// Kanban shows a kanban swimlane view of a collection.
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
		milestoneID := getInt(r.URL, "milestone", 0)

		p, err := h.collectionPage(r.Context(), id, isRefresh(r))
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("page: %v", err)

			return
		}

		chosen, milestones := milestoneChoices(p.CollectionResult.RuleResults, milestoneID)
		klog.Infof("milestones choices: %+v", milestones)

		p.Description = p.Collection.Description
		p.Swimlanes = groupByUser(p.CollectionResult.RuleResults, chosen.GetNumber())
		p.SelectorOptions = milestones
		p.SelectorVar = "milestone"
		p.Milestone = chosen

		klog.V(2).Infof("page context: %+v", p)

		err = t.ExecuteTemplate(w, "base", p)
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("tmpl: %v", err)

			return
		}
	}
}

func milestoneChoices(results []*triage.RuleResult, milestoneID int) (*github.Milestone, []Choice) {
	mmap := map[int]*github.Milestone{}

	for _, r := range results {
		for _, co := range r.Items {
			mmap[co.Milestone.GetNumber()] = co.Milestone
		}
	}

	milestones := []*github.Milestone{}
	for _, v := range mmap {
		milestones = append(milestones, v)
	}

	sort.Slice(milestones, func(i, j int) bool { return milestones[i].GetDueOn().After(milestones[j].GetDueOn()) })

	if milestoneID == 0 {
		milestoneID = milestones[0].GetNumber()
	}

	choices := []Choice{}

	var chosen *github.Milestone

	for _, m := range milestones {
		c := Choice{
			Value: m.GetNumber(),
			Text:  fmt.Sprintf("%s (%s)", m.GetTitle(), m.GetDueOn().Format("2006-01-02")),
		}
		if c.Value == milestoneID {
			c.Selected = true
			chosen = m
		}

		choices = append(choices, c)
	}

	return chosen, choices
}
