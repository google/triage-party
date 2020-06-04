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
	"context"
	"fmt"
	"time"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

func (h *Handlers) collectionPage(ctx context.Context, id string, refresh bool) (*Page, error) {
	start := time.Now()
	dataAge := time.Time{}

	defer func() {
		klog.Infof("Served %q request within %s from data %s old", id, time.Since(start), time.Since(dataAge))
	}()

	s, err := h.party.LookupCollection(id)
	if err != nil {
		return nil, fmt.Errorf("lookup collection: %w", err)
	}

	sts, err := h.party.ListCollections()
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}

	var result *triage.CollectionResult
	if refresh {
		result = h.updater.ForceRefresh(ctx, id)
		klog.Infof("refresh %q result: %d items", id, len(result.RuleResults))
	} else {
		result = h.updater.Lookup(ctx, id, true)
		if result == nil {
			return nil, fmt.Errorf("lookup %q returned no data", id)
		}

		if result.RuleResults == nil {
			return nil, fmt.Errorf("lookup %q returned no results", id)
		}

		klog.V(2).Infof("lookup %q result: %d items", id, len(result.RuleResults))
	}

	dataAge = result.Time
	warning := ""

	if time.Since(result.Time) > h.warnAge {
		warning = fmt.Sprintf("Serving results from %s ago. Service started %s ago and is downloading new data. Use Shift-Reload to force refresh at any time.", humanDuration(time.Since(result.Time)), humanDuration(time.Since(h.startTime)))
	}

	total := 0
	for _, o := range result.RuleResults {
		total += len(o.Items)
	}

	uniqueFiltered := []*hubbub.Conversation{}
	seenFiltered := map[int]bool{}

	for _, o := range result.RuleResults {
		for _, i := range o.Items {
			if !seenFiltered[i.ID] {
				uniqueFiltered = append(uniqueFiltered, i)
				seenFiltered[i.ID] = true
			}
		}
	}

	unique := []*hubbub.Conversation{}
	seen := map[int]bool{}

	for _, o := range result.RuleResults {
		for _, i := range o.Items {
			if !seen[i.ID] {
				unique = append(unique, i)
				seen[i.ID] = true
			}
		}
	}

	p := &Page{
		ID:               s.ID,
		Version:          VERSION,
		SiteName:         h.siteName,
		Title:            s.Name,
		Collection:       s,
		Collections:      sts,
		Description:      s.Description,
		CollectionResult: result,
		Total:            len(unique),
		TotalShown:       len(uniqueFiltered),
		Types:            "Issues",
		Warning:          warning,
		UniqueItems:      uniqueFiltered,
		ResultAge:        time.Since(result.Time),
	}

	for _, s := range sts {
		if s.UsedForStats {
			p.Stats = h.updater.Lookup(ctx, s.ID, false)
			p.StatsID = s.ID
		}
	}

	return p, nil
}
