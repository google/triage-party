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

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

func (h *Engine) cachedOrgMembers(ctx context.Context, org string) (map[string]bool, error) {
	key := fmt.Sprintf("%s-members", org)
	if x, ok := h.cache.Get(key); ok {
		members := x.(map[string]bool)
		return members, nil
	}
	klog.V(1).Infof("members miss: %s", key)
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

	h.cache.Set(key, members, h.maxEventAge)
	klog.Infof("%s has %d members", org, len(members))
	return members, nil
}
