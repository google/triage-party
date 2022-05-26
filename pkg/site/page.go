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
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

const (
	// OpenStatsName is the name of the rule containing open items stats
	OpenStatsName = "__open__"
	// VelocityStatsName is the name of the rulee containing velocity stats
	VelocityStatsName = "__velocity__"
)

func (h *Handlers) collectionPage(ctx context.Context, id string, refresh bool) (*Page, error) {
	start := time.Now()

	defer func() {
		klog.Infof("Served %q request within %s", id, time.Since(start))
	}()

	s, err := h.party.LookupCollection(id)
	if err != nil {
		return nil, fmt.Errorf("lookup collection: %w", err)
	}

	// The input ID may have been a collection ID or category.  LookupCollection
	// will return the first collection for a category ID.  Normalize this now to
	// the chosen collection ID.
	id = s.ID

	sts, err := h.party.ListCollections()
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}

	categories, err := h.party.ListCategories()
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	var result *triage.CollectionResult
	if refresh {
		result = h.updater.ForceRefresh(ctx, id)
		klog.Infof("refresh %q result: %d items", id, len(result.RuleResults))
	} else {
		result = h.updater.Lookup(ctx, id, false)
		if result == nil {
			klog.Errorf("lookup %q returned no data", id)
			result = &triage.CollectionResult{}
		} else if result.RuleResults == nil {
			klog.Errorf("lookup %q returned no results: %+v", id, result)
		}
	}

	total := 0
	for _, o := range result.RuleResults {
		total += len(o.Items)
	}

	unique := uniqueItems(result.RuleResults)

	p := &Page{
		ID:               s.ID,
		Category:         s.Category,
		Version:          VERSION,
		SiteName:         h.siteName,
		Title:            s.Name,
		Collection:       s,
		Collections:      sts,
		Categories:       categories,
		Description:      s.Description,
		CollectionResult: result,
		Total:            len(unique),
		Types:            "Issues",
		UniqueItems:      unique,
		ResultAge:        time.Since(result.OldestInput),
		Status:           h.updater.Status(),
	}

	if result.RuleResults == nil {
		p.Notification = template.HTML(fmt.Sprintf("No cached data found - performing initial data download (%d issues examined) ...", h.party.ConversationsTotal()))
	} else if p.ResultAge > h.warnAge {
		p.Notification = template.HTML(fmt.Sprintf(`Refreshing data in the background. Displayed data may be up to %s old. Use <a href="https://en.wikipedia.org/wiki/Wikipedia:Bypass_your_cache#Bypassing_cache">Shift-Reload</a> to force a data refresh at any time.`, humanDuration(time.Since(result.OldestInput))))
		p.Stale = true
	}

	if result.Collection != nil && result.Collection.Velocity != "" {
		p.VelocityStats = h.updater.Lookup(ctx, result.Collection.Velocity, false)
	} else {
		for _, s := range sts {
			if s.UsedForStats {
				if s.ID == VelocityStatsName {
					p.VelocityStats = h.updater.Lookup(ctx, s.ID, false)
					continue
				}
				// Older configs may not use OpenStatsName
				if s.ID == OpenStatsName || p.OpenStats == nil {
					p.OpenStats = h.updater.Lookup(ctx, s.ID, false)
					continue
				}
			}
		}
	}

	return p, nil
}

func uniqueItems(results []*triage.RuleResult) []*hubbub.Conversation {
	items := []*hubbub.Conversation{}
	seen := map[string]bool{}

	for _, r := range results {
		for _, i := range r.Items {
			if !seen[i.URL] {
				seen[i.URL] = true
				items = append(items, i)
			}
		}
	}
	return items
}
