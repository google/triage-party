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
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// mtime is a workaround the GitHub misfeature that UpdatedAt is not incremented for cross-reference events
func (h *Engine) mtime(i GitHubItem) time.Time {
	updatedAt := i.GetUpdatedAt()
	key := updateKey(i)
	updateSeen := h.updatedAt[key]

	if updateSeen == updatedAt {
		return updatedAt
	}

	if updateSeen.After(updatedAt) {
		klog.V(1).Infof("%s has updates from %s, after last update %s", key, updateSeen, updatedAt)
		return updateSeen
	} else if !updatedAt.IsZero() {
		klog.V(3).Infof("%s has updates from %s, before last update %s", key, updateSeen, updatedAt)
	}

	return updatedAt
}

// mtimeRef is like mtime, but for related conversations
func (h *Engine) mtimeRef(rc *RelatedConversation) time.Time {
	updatedAt := rc.Updated
	key := fmt.Sprintf("%s/%s#%d", rc.Organization, rc.Project, rc.ID)
	updateSeen := h.updatedAt[key]

	if updateSeen == updatedAt {
		return updatedAt
	}

	if updateSeen.After(updatedAt) {
		klog.V(1).Infof("%s has updates from %s, after last update %s", key, updateSeen, updatedAt)
		return updateSeen
	} else if !updatedAt.IsZero() {
		klog.V(3).Infof("%s has updates from %s, before last update %s", key, updateSeen, updatedAt)
	}

	return updatedAt
}

func updateKey(i GitHubItem) string {
	// https://github.com/kubernetes/minikube/pull/8431
	parts := strings.Split(i.GetHTMLURL(), "/")
	num := parts[len(parts)-1]
	project := parts[len(parts)-3]
	org := parts[len(parts)-4]
	return fmt.Sprintf("%s/%s#%s", org, project, num)
}

func (h *Engine) updateMtime(i GitHubItem) {
	key := updateKey(i)
	h.updateMtimeByKey(key, i.GetUpdatedAt())
}

func (h *Engine) updateMtimeLong(org string, project string, num int, ts time.Time) {
	key := fmt.Sprintf("%s/%s#%d", org, project, num)
	h.updateMtimeByKey(key, ts)
}

func (h *Engine) updateMtimeByKey(key string, ts time.Time) {
	if ts.After(h.updatedAt[key]) {
		klog.V(3).Infof("Updating %s last update time for %s to %s", key, h.updatedAt[key], ts)
		h.updatedAt[key] = ts
	}
}
