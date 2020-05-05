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
	"github.com/google/triage-party/pkg/logu"
	"github.com/google/triage-party/pkg/persist"
	"k8s.io/klog/v2"
)

func (h *Engine) cachedOrgMembers(ctx context.Context, org string, newerThan time.Time) (map[string]bool, error) {
	key := fmt.Sprintf("%s-members", org)

	if x := h.cache.GetNewerThan(key, newerThan); x != nil {
		return x.StringBool, nil
	}

	klog.Infof("members miss: %s newer than %s", key, logu.STime(newerThan))
	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	members := map[string]bool{}
	for {
		klog.Infof("Downloading members of %q org (page %d)...", org, opt.Page)
		mem, resp, err := h.client.Organizations.ListMembers(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, m := range mem {
			members[m.GetLogin()] = true
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err := h.cache.Set(key, &persist.Thing{StringBool: members}); err != nil {
		klog.Errorf("set %q failed: %v", key, err)
	}

	klog.Infof("%s has %d members", org, len(members))
	return members, nil
}
