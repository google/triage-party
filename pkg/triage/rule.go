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

package triage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/provider"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/logu"
	"k8s.io/klog/v2"
)

// Rule is a logical triage group
type Rule struct {
	ID         string
	Resolution string            `yaml:"resolution,omitempty"`
	Name       string            `yaml:"name,omitempty"`
	Repos      []string          `yaml:"repos,omitempty"`
	Type       string            `yaml:"type,omitempty"`
	Filters    []provider.Filter `yaml:"filters"`
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

	Duplicates map[string]bool

	// OldestInput is the timestamp of the oldest input data
	OldestInput time.Time

	// When was this rule result created?
	Created time.Time
}

// SummarizeRuleResult adds together statistics about a pool of conversations
func SummarizeRuleResult(t Rule, cs []*hubbub.Conversation, seen map[string]*Rule) *RuleResult {
	r := &RuleResult{
		Rule:       t,
		Duplicates: map[string]bool{},
	}

	if seen == nil {
		r.Items = cs
	} else {
		for i, c := range cs {
			if c == nil {
				klog.Errorf("SummarizeRuleResult conversation #%d is nil: %+v", i, c)
				continue
			}

			dupeRule := seen[c.URL]
			if dupeRule != nil {
				// Find a nefarious bug
				if t.ID == dupeRule.ID {
					klog.Errorf("LOGIC FAILURE: %s was previously seen by rule %q, which is the same as the current rule: %q", c.URL, dupeRule.ID, t.ID)
					continue
				}

				klog.V(2).Infof("dupe: %s (now: %q, previous: %q)", c.URL, t.ID, dupeRule.ID)
				r.Duplicates[c.URL] = true
			}
			r.Items = append(r.Items, c)
			seen[c.URL] = &t
		}
	}

	if len(cs) == 0 {
		return r
	}

	for i, c := range cs {
		if c == nil {
			klog.Errorf("SummarizeRuleResult conversation #%d is nil: %+v", i, c)
			continue
		}

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
	r.Created = time.Now()
	return r
}

// ExecuteRule executes a rule. seen is optional.
func (p *Party) ExecuteRule(ctx context.Context, sp provider.SearchParams, t Rule, seen map[string]*Rule, s *Collection) (*RuleResult, error) {
	klog.V(1).Infof("executing rule %q for results newer than %s", t.ID, logu.STime(sp.NewerThan))
	rcs := []*hubbub.Conversation{}
	oldest := time.Now()

	if s != nil {
		if len(t.Repos) == 0 && len(s.Repos) > 0 {
			t.Repos = s.Repos
		}
	}

	for _, repoUrl := range t.Repos {
		r, err := parseRepo(repoUrl)
		if err != nil {
			return nil, err
		}

		klog.V(2).Infof("%s -> org=%s project=%s", repoUrl, r.Organization, r.Project)

		var ts time.Time
		var cs []*hubbub.Conversation

		sp.Repo = r
		sp.Filters = t.Filters

		switch t.Type {
		case hubbub.Issue:
			cs, ts, err = p.engine.SearchIssues(ctx, sp)
		case hubbub.PullRequest:
			cs, ts, err = p.engine.SearchPullRequests(ctx, sp)
		default:
			cs, ts, err = p.engine.SearchAny(ctx, sp)
		}

		if err != nil {
			return nil, err
		}

		rcs = append(rcs, cs...)
		if ts.Before(oldest) {
			oldest = ts
		}
	}

	klog.V(1).Infof("rule %q matched %d items", t.ID, len(rcs))
	rr := SummarizeRuleResult(t, rcs, seen)
	rr.OldestInput = oldest
	return rr, nil
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

// parseRepo returns provider, organization and project for a URL
// rawURL should be a valid url with host like https://github.com/org/repo
// or https://gitlab.com/org/repo
// or https://gitlab.com/org/group/repo
func parseRepo(rawURL string) (r provider.Repo, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if u.Host == "" {
		err = fmt.Errorf("Provided string %s is not a valid URL", rawURL)
		return
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 && len(parts) != 4 {
		// gitlab may have https://gitlab.com/organization/group/repo
		err = fmt.Errorf("expected 2/3 repository parts, got %d: %v", len(parts), parts)
		return
	}
	if len(parts) == 3 {
		r = provider.Repo{
			Host:         u.Host,
			Organization: parts[1],
			Project:      parts[2],
		}
	} else {
		r = provider.Repo{
			Host:         u.Host,
			Organization: parts[1],
			Group:        parts[2],
			Project:      parts[3],
		}
	}

	return
}
