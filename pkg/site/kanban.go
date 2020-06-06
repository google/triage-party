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
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

var unassigned = "zz_unassigned"

// veryLate is how long before a milestone is considered very late
var veryLate = 24 * 6 * time.Hour

// Swimlane is a row in a Kanban display.
type Swimlane struct {
	User    *github.User
	Columns []*triage.RuleResult
}

func avatarWide(u *github.User) template.HTML {
	if u.GetLogin() == unassigned {
		return template.HTML(`<div class="unassigned"><div class="unassigned-icon" title="Unassigned work - free for the taking!"></div><span>nobody</span></div>`)
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

		if len(p.CollectionResult.RuleResults) == 0 {
			http.Error(w, fmt.Sprintf("no results for %q", id), 400)
			return
		}

		chosen, milestones := milestoneChoices(p.CollectionResult.RuleResults, milestoneID)
		klog.Infof("milestones choices: %+v", milestones)

		p.Description = p.Collection.Description
		p.Swimlanes = groupByUser(p.CollectionResult.RuleResults, chosen.GetNumber())
		p.SelectorOptions = milestones
		p.SelectorVar = "milestone"
		p.Milestone = chosen
		p.ClosedPerDay = calcClosedPerDay(p.VelocityStats)
		p.MilestoneETA = calcETA(chosen, p.ClosedPerDay)

		dueOn := p.Milestone.GetDueOn()

		if p.MilestoneETA.Format("2006-01-02") == dueOn.Format("2006-01-02") {
			klog.Infof("milestone ETA is on time!")
			p.MilestoneOnTarget = 0
		} else if p.MilestoneETA.After(dueOn) {
			p.MilestoneOnTarget = 1
			p.MilestoneETADiff = p.MilestoneETA.Sub(dueOn)
			if p.MilestoneETADiff > veryLate {
				p.MilestoneOnTarget = 2
			}
		} else {
			p.MilestoneOnTarget = -1
			p.MilestoneETADiff = dueOn.Sub(p.MilestoneETA)
		}

		klog.V(2).Infof("page context: %+v", p)

		err = t.ExecuteTemplate(w, "base", p)
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}

func calcClosedPerDay(r *triage.CollectionResult) float64 {
	if r == nil {
		klog.Errorf("unable to calc closed per day: no data")
		return 0.0
	}

	oldestClosure := time.Now()

	for _, r := range r.RuleResults {
		for _, co := range r.Items {
			if !co.ClosedAt.IsZero() && co.ClosedAt.Before(oldestClosure) {
				klog.Infof("#%d was closed at %s", co.ID, co.ClosedAt)
				oldestClosure = co.ClosedAt
			}
		}
	}

	days := time.Since(oldestClosure).Hours() / 24
	closeRate := days / float64(r.TotalIssues)
	klog.Infof("close rate is %.2f (%.1f days of data, %d issues)", closeRate, days, r.TotalIssues)
	return closeRate
}

func calcETA(m *github.Milestone, closeRate float64) time.Time {
	if m == nil {
		klog.Errorf("unable to calc ETA: no milestone")
		return time.Time{}
	}

	open := m.GetOpenIssues()

	if open == 0 {
		klog.Errorf("unable to calc ETA: no issues")
		return time.Time{}
	}

	if closeRate < 0.0001 {
		klog.Errorf("unable to calc ETA: too low of a close rate: %f", closeRate)
		return time.Time{}
	}

	days := float64(open) / closeRate
	return time.Now().AddDate(0, 0, int(days))
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

	if len(milestones) == 0 {
		klog.Errorf("Went through %d issues, but none had a milestone", len(results))
		return nil, nil
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
