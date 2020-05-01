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

package triage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/logu"
	"k8s.io/klog/v2"
)

// Rule is a logical triage group
type Rule struct {
	ID         string
	Resolution string          `yaml:"resolution,omitempty"`
	Name       string          `yaml:"name,omitempty"`
	Repos      []string        `yaml:"repos,omitempty"`
	Type       string          `yaml:"type,omitempty"`
	Filters    []hubbub.Filter `yaml:"filters"`
}

type RuleResult struct {
	Rule  Rule
	Items []*hubbub.Conversation

	AvgAge             time.Duration
	AvgCurrentHold     time.Duration
	AvgAccumulatedHold time.Duration

	// Avoiding time.Duration because it's easy to int64 overflow
	TotalAgeDays             float64
	TotalCurrentHoldDays     float64
	TotalAccumulatedHoldDays float64

	Duplicates int
}

// SummarizeRuleResult adds together statistics about a pool of conversations
func SummarizeRuleResult(t Rule, cs []*hubbub.Conversation, seen map[string]*Rule) *RuleResult {
	klog.V(2).Infof("Summarizing %q with %d conversations, seen has %d members", t.ID, len(cs), len(seen))

	r := &RuleResult{
		Rule:       t,
		Duplicates: 0,
	}

	if seen == nil {
		r.Items = cs
	} else {
		for _, c := range cs {
			dupeRule := seen[c.URL]
			if dupeRule != nil {
				// Find a nefarious bug
				if t.ID == dupeRule.ID {
					panic(fmt.Sprintf("this can't happen: %s is a double-dupe of %s", c.URL, dupeRule.ID))
				}

				klog.V(2).Infof("dupe: %s (now: %q, previous: %q)", c.URL, t.ID, dupeRule.ID)
				c.Hidden = true
				r.Duplicates++
			}
			r.Items = append(r.Items, c)
			seen[c.URL] = &t
		}
	}

	if len(cs) == 0 {
		return r
	}

	for _, c := range cs {
		if c.Created.After(time.Now()) {
			klog.Errorf("#%d claims to have be newer than now: %s", c.ID, c.Created)
			continue
		}
		r.TotalAgeDays += time.Since(c.Created).Hours() / 24
		r.TotalCurrentHoldDays += c.CurrentHoldTime.Hours() / 24
		r.TotalAccumulatedHoldDays += c.AccumulatedHoldTime.Hours() / 24
	}

	count := len(cs)

	r.AvgAge = avgDayDuration(r.TotalAgeDays, count)
	r.AvgCurrentHold = avgDayDuration(r.TotalCurrentHoldDays, count)
	r.AvgAccumulatedHold = avgDayDuration(r.TotalAccumulatedHoldDays, count)
	return r
}

// ExecuteRule executes a rule. seen is optional.
func (p *Party) ExecuteRule(ctx context.Context, t Rule, seen map[string]*Rule, newerThan time.Time) (*RuleResult, error) {
	klog.Infof("executing rule %q for results newer than %s (stale_ok=%v)", t.ID, logu.STime(newerThan), p.acceptStaleResults)
	rcs := []*hubbub.Conversation{}

	for _, repo := range t.Repos {
		org, project, err := parseRepo(repo)
		if err != nil {
			return nil, err
		}

		klog.V(2).Infof("%s -> org=%s project=%s", repo, org, project)

		var cs []*hubbub.Conversation
		switch t.Type {
		case hubbub.Issue:

			cs, err = p.engine.SearchIssues(ctx, org, project, t.Filters, newerThan)
		case hubbub.PullRequest:
			cs, err = p.engine.SearchPullRequests(ctx, org, project, t.Filters, newerThan)
		default:
			cs, err = p.engine.SearchAny(ctx, org, project, t.Filters, newerThan)
		}

		if err != nil {
			return nil, err
		}

		rcs = append(rcs, cs...)
	}

	klog.V(1).Infof("rule %q matched %d items", t.ID, len(rcs))
	return SummarizeRuleResult(t, rcs, seen), nil
}

// Return a fully resolved rule
func (p *Party) LookupRule(id string) (Rule, error) {
	t, ok := p.rules[id]
	if !ok {
		return t, fmt.Errorf("rule %q is undefined - typo?", id)
	}
	t.ID = id
	if len(p.reposOverride) > 0 {
		t.Repos = p.reposOverride
	}

	if len(t.Repos) == 0 {
		t.Repos = p.settings.Repos
	}
	return t, nil
}

// ListRules fully resolved rules
func (p *Party) ListRules() ([]Rule, error) {
	ts := []Rule{}
	for k := range p.rules {
		s, err := p.LookupRule(k)
		if err != nil {
			return ts, err
		}
		ts = append(ts, s)
	}
	return ts, nil
}
