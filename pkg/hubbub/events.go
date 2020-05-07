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
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/persist"
	"k8s.io/klog/v2"
)

func (h *Engine) cachedTimeline(ctx context.Context, org string, project string, num int, newerThan time.Time) ([]*github.Timeline, error) {
	key := fmt.Sprintf("%s-%s-%d-timeline", org, project, num)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.Timeline, nil
	}

	klog.V(1).Infof("cache miss for %s newer than %s", key, newerThan)
	return h.updateTimeline(ctx, org, project, num, key)
}

func (h *Engine) updateTimeline(ctx context.Context, org string, project string, num int, key string) ([]*github.Timeline, error) {
	klog.V(1).Infof("Downloading timeline for %s/%s #%d", org, project, num)

	opt := &github.ListOptions{
		PerPage: 100,
	}
	var allEvents []*github.Timeline
	for {
		klog.V(2).Infof("Downloading timeline for %s/%s #%d (page %d)...", org, project, num, opt.Page)
		evs, resp, err := h.client.Issues.ListIssueTimeline(ctx, org, project, num, opt)
		if err != nil {
			return nil, err
		}
		h.logRate(resp.Rate)

		klog.V(2).Infof("Received %d timeline events", len(evs))
		allEvents = append(allEvents, evs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{Timeline: allEvents}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	return allEvents, nil
}
