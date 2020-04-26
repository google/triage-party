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

package hubbub

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/imjasonmiller/godice"
	"k8s.io/klog"
)

const Issue = "issue"
const PullRequest = "pull_request"
const MinSeenForSimilarity = 5

// The result of Execute
type CollectionResult struct {
	Time        time.Time
	RuleResults []RuleResult

	Total             int
	TotalPullRequests int
	TotalIssues       int

	AvgHold  time.Duration
	AvgAge   time.Duration
	AvgDelay time.Duration

	TotalHold  time.Duration
	TotalAge   time.Duration
	TotalDelay time.Duration
}

type RuleResult struct {
	Rule  Rule
	Items []*Colloquy

	AvgHold  time.Duration
	AvgAge   time.Duration
	AvgDelay time.Duration

	TotalHold  time.Duration
	TotalAge   time.Duration
	TotalDelay time.Duration

	Duplicates int
}

// ExecuteCollection executes a collection.
func (h *HubBub) ExecuteCollection(ctx context.Context, client *github.Client, s Collection) (*CollectionResult, error) {
	klog.Infof("executing collection %q", s.ID)
	os := []RuleResult{}
	seen := map[int]bool{}

	for _, tid := range s.RuleIDs {
		t, err := h.LookupRule(tid)
		if err != nil {
			return nil, err
		}

		cs, err := h.ExecuteRule(ctx, client, t)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %v", t.Name, err)
		}
		os = append(os, SummarizeRuleResult(t, cs, seen))
	}

	r := SummarizeCollectionResult(os)
	r.Time = time.Now()
	return r, nil
}

// SummarizeCollectionResult adds together statistics about collection results {
func SummarizeCollectionResult(os []RuleResult) *CollectionResult {
	r := &CollectionResult{}
	for _, oc := range os {
		r.Total += len(oc.Items)
		if oc.Rule.Type == PullRequest {
			r.TotalPullRequests += len(oc.Items)
		} else {
			r.TotalIssues += len(oc.Items)
		}
		r.RuleResults = append(r.RuleResults, oc)
		r.TotalHold += oc.TotalHold
		r.TotalAge += oc.TotalAge
		r.TotalDelay += oc.TotalDelay

	}
	if r.Total > 0 {
		r.AvgHold = time.Duration(int64(r.TotalHold) / int64(r.Total))
		r.AvgAge = time.Duration(int64(r.TotalAge) / int64(r.Total))
		r.AvgDelay = time.Duration(int64(r.TotalDelay) / int64(r.Total))
	}
	return r
}

func (h *HubBub) similar(c *Colloquy) ([]int, error) {
	if len(h.seen) < MinSeenForSimilarity {
		return nil, nil
	}
	min := h.settings.MinSimilarity
	if min == 0 {
		return nil, nil
	}
	// We should measure if caching is worth it, and if so, pick a better key.
	key := fmt.Sprintf("similar-v2-%.2f-%d-%s", min, len(h.seen), c.Title)
	if x, ok := h.cache.Get(key); ok {
		similar := x.([]int)
		return similar, nil
	}
	choices := []string{}
	for id, sc := range h.seen {
		if id == c.ID {
			continue
		}
		if c.Type == sc.Type {
			choices = append(choices, sc.Title)
		}
	}

	matches, err := godice.CompareStrings(c.Title, choices)
	if err != nil {
		return nil, err
	}

	var similar []int
	for _, match := range matches.Candidates {
		if match.Score > min {
			similar = append(similar, h.seenTitles[match.Text])
		}
	}

	h.cache.Set(key, similar, h.maxEventAge)
	return similar, nil
}

// SummarizeRuleResult adds together statistics about a pool of conversations
func SummarizeRuleResult(t Rule, cs []*Colloquy, dedup map[int]bool) RuleResult {
	r := RuleResult{
		Rule:       t,
		Duplicates: 0,
	}

	if dedup == nil {
		r.Items = cs
	} else {
		for _, c := range cs {
			if dedup[c.ID] {
				c.Hidden = true
				r.Duplicates++
			}
			r.Items = append(r.Items, c)
			dedup[c.ID] = true
		}
	}

	if len(cs) == 0 {
		return r
	}

	for _, c := range cs {
		r.TotalDelay += c.LatestResponseDelay
		r.TotalHold += time.Since(c.OnHoldSince)
		r.TotalAge += time.Since(c.Created)
	}

	count := int64(len(cs))
	r.AvgHold = time.Duration(int64(r.TotalHold) / count)
	r.AvgAge = time.Duration(int64(r.TotalAge) / count)
	r.AvgDelay = time.Duration(int64(r.TotalDelay) / count)
	return r
}

// ExecuteRule executes a rule.
func (h *HubBub) ExecuteRule(ctx context.Context, client *github.Client, t Rule) ([]*Colloquy, error) {
	klog.Infof("executing rule %q", t.ID)
	result := []*Colloquy{}

	for _, repo := range t.Repos {
		org, project, err := parseRepo(repo)
		klog.V(2).Infof("%s -> org=%s project=%s", repo, org, project)

		if err != nil {
			return result, err
		}

		var cs []*Colloquy
		switch t.Type {
		case Issue:
			cs, err = h.Issues(ctx, org, project, t.Filters)
		case PullRequest:
			cs, err = h.PullRequests(ctx, org, project, t.Filters)
		default:
			cs, err = h.Issues(ctx, org, project, t.Filters)
			if err != nil {
				return result, err
			}
			pcs, err := h.PullRequests(ctx, org, project, t.Filters)
			if err != nil {
				return result, err
			}
			cs = append(cs, pcs...)
		}
		if err != nil {
			return result, err
		}
		for _, c := range cs {
			h.seen[c.ID] = c
			h.seenTitles[c.Title] = c.ID
		}
		result = append(result, cs...)
	}

	for _, c := range result {
		sim, err := h.similar(c)
		if err != nil {
			klog.Errorf("unable to find similar for %d: %v", c.ID, err)
			continue
		}
		if len(sim) > 0 {
			c.Similar = []RelatedColloquy{}
			for _, id := range sim {
				if h.seen[id] == nil {
					klog.Errorf("have not seen related item: %d", id)
					continue
				}
				c.Similar = append(c.Similar, RelatedColloquy{
					ID:      id,
					URL:     h.seen[id].URL,
					Title:   h.seen[id].Title,
					Author:  h.seen[id].Author,
					Type:    h.seen[id].Type,
					Created: h.seen[id].Created,
				})
			}
		}
	}
	return result, nil
}

// parseRepo returns the organization and project for a URL
func parseRepo(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(u.Path, "/")

	// not a URL
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	// URL
	if len(parts) == 3 {
		return parts[1], parts[2], nil
	}
	return "", "", fmt.Errorf("expected 2 repository parts, got %d: %v", len(parts), parts)
}
