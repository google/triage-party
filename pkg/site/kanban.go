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
	"math"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/provider"

	"github.com/dustin/go-humanize"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

var unassigned = "zz_unassigned"

// Swimlane is a row in a Kanban display.
type Swimlane struct {
	User    *provider.User
	Columns []*triage.RuleResult
	Issues  int
}

func avatarWide(u *provider.User) template.HTML {
	if u.GetLogin() == unassigned {
		return template.HTML(`<div class="unassigned"><div class="unassigned-icon" title="Unassigned work - free for the taking!"></div><span>nobody</span></div>`)
	}

	return template.HTML(fmt.Sprintf(`<a href="%s" title="%s"><img src="%s" width="96" height="96"></a>`, u.GetHTMLURL(), u.GetLogin(), u.GetAvatarURL()))
}

func groupByUser(results []*triage.RuleResult, milestoneID int, dedup bool) []*Swimlane {
	lanes := map[string]*Swimlane{}
	seenItem := map[string]bool{}

	for i, r := range results {
		for _, co := range r.Items {
			if milestoneID > 0 && co.Milestone.GetNumber() != milestoneID {
				continue
			}

			assignees := co.Assignees
			if len(assignees) == 0 {
				assignees = append(assignees, &provider.User{
					Login: &unassigned,
				})
			}

			for _, a := range assignees {
				// Dedup across users and columns
				if dedup && seenItem[co.URL] {
					continue
				}

				seenItem[co.URL] = true

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
				lanes[assignee].Issues++
			}
		}
	}

	ls := []*Swimlane{}
	for _, v := range lanes {
		ls = append(ls, v)
	}

	return ls
}

func lateTime(t time.Time, ref time.Time) string {
	return humanize.CustomRelTime(t, ref, "early", "late", defaultMagnitudes)
}

// Kanban shows a kanban swimlane view of a collection.
func (h *Handlers) Kanban() http.HandlerFunc {
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
	}

	t := template.Must(template.New("kanban").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "kanban.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/k/")
		milestoneID := getInt(r.URL, "milestone", -1)

		p, err := h.collectionPage(r.Context(), id, isRefresh(r))
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("page: %v", err)
			return
		}

		if p.CollectionResult.RuleResults != nil {
			chosen, milestones := milestoneChoices(p.CollectionResult.RuleResults, milestoneID)
			klog.Infof("milestones chosen: %d, choices: %+v", milestoneID, milestones)

			p.Description = p.Collection.Description
			p.Swimlanes = groupByUser(p.CollectionResult.RuleResults, chosen.GetNumber(), p.Collection.Dedup)
			p.SelectorOptions = milestones
			p.SelectorVar = "milestone"
			p.Milestone = chosen
			p.ClosedPerDay = calcClosedPerDay(p.VelocityStats)
			p.CompletionETA = calcETA(p.Swimlanes, p.ClosedPerDay)

			etaDate, etaOffset, countOffset := calcMilestoneETA(chosen, p.ClosedPerDay)
			klog.Infof("milestone ETA is %s (offset: %s, %d issues)", etaDate, etaOffset, countOffset)
			p.MilestoneETA = etaDate
			p.MilestoneCountOffset = countOffset

			if etaOffset > 6*24*time.Hour {
				p.MilestoneVeryLate = true
			}
		}

		err = t.ExecuteTemplate(w, "base", p)
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}

func calcETA(lanes []*Swimlane, perDay float64) time.Time {
	open := map[string]bool{}

	for _, lane := range lanes {
		for _, c := range lane.Columns {
			if c != nil {
				for _, co := range c.Items {
					if (co.State == constants.OpenState) || (co.State == constants.OpenedState) {
						open[co.URL] = true
					}
				}
			}
		}
	}

	days := float64(len(open)) / perDay
	return time.Now().AddDate(0, 0, int(days))
}

func calcClosedPerDay(r *triage.CollectionResult) float64 {
	if r == nil {
		klog.Errorf("unable to calc closed per day: no data")
		return 0.0
	}

	oldestClosure := time.Now()
	// dedup
	seen := map[string]bool{}

	for _, r := range r.RuleResults {
		for _, co := range r.Items {
			if seen[co.URL] {
				continue
			}
			seen[co.URL] = true

			if !co.ClosedAt.IsZero() && co.ClosedAt.Before(oldestClosure) {
				klog.V(1).Infof("#%d was closed at %s", co.ID, co.ClosedAt)
				oldestClosure = co.ClosedAt
			}
		}
	}

	days := time.Since(oldestClosure).Hours() / 24
	closeRate := float64(len(seen)) / days
	klog.Infof("close rate is %.2f (%.1f days of data, %d issues)", closeRate, days, r.TotalIssues)
	return closeRate
}

// TODO: Merge into calcETA
func calcMilestoneETA(m *provider.Milestone, closeRate float64) (time.Time, time.Duration, int) {
	if m == nil {
		klog.Errorf("unable to calc ETA: no milestone")
		return time.Time{}, time.Duration(0), 0
	}

	open := m.GetOpenIssues()

	if open == 0 {
		klog.Errorf("unable to calc ETA: no issues")
		return time.Time{}, time.Duration(0), 0
	}

	if closeRate < 0.0001 {
		klog.Errorf("unable to calc ETA: too low of a close rate: %f", closeRate)
		return time.Time{}, time.Duration(0), 0
	}

	// How many will we get done by the due date?
	daysToDue := m.GetDueOn().Sub(time.Now()).Hours() / 24
	canShip := daysToDue * closeRate
	klog.Errorf("%.2f days until due date, can ship %.2f items", daysToDue, canShip)

	days := float64(open) / closeRate
	eta := time.Now().AddDate(0, 0, int(days))

	overByDuration := eta.Sub(m.GetDueOn())
	overByCount := int(math.Ceil(float64(open) - canShip))
	return eta, overByDuration, overByCount
}

func milestoneChoices(results []*triage.RuleResult, milestoneID int) (*provider.Milestone, []Choice) {
	mmap := map[int]*provider.Milestone{}

	notInMilestone := 0

	for _, r := range results {
		for _, co := range r.Items {
			if co.Milestone == nil || co.Milestone.GetNumber() == 0 {
				if notInMilestone == 0 {
					klog.Infof("Found issue within %s that is not in a milestone: %s", r.Rule.ID, co.URL)
				}
				notInMilestone++
				continue
			}
			mmap[co.Milestone.GetNumber()] = co.Milestone
		}
	}

	milestones := []*provider.Milestone{}
	for _, v := range mmap {
		milestones = append(milestones, v)
	}

	if len(milestones) == 0 {
		klog.Errorf("Went through %d issues, but none had a milestone", len(results))
		return nil, nil
	}

	sort.Slice(milestones, func(i, j int) bool { return milestones[i].GetDueOn().Before(milestones[j].GetDueOn()) })

	// Only auto-select a milestone if all issues are within a milestone
	if milestoneID == -1 {
		if len(milestones) > 0 && notInMilestone == 0 {
			if c := currentMilestone(milestones); c != nil {
				milestoneID = c.GetNumber()
			} else {
				milestoneID = milestones[0].GetNumber()
			}
		} else {
			milestoneID = 0 // all
		}
	}

	choices := []Choice{}

	var chosen *provider.Milestone

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

	choices = append(choices, Choice{
		Value:    0,
		Text:     "All items",
		Selected: milestoneID == 0,
	})

	return chosen, choices
}

func currentMilestone(milestones []*github.Milestone) *github.Milestone {
	var curr *github.Milestone
	today := time.Now()
	for _, m := range milestones {
		if m.GetDueOn().Before(today) {
			continue
		}
		if curr == nil {
			curr = m
		} else if m.GetDueOn().Before(curr.GetDueOn()) {
			curr = m
		}
	}
	return curr
}
