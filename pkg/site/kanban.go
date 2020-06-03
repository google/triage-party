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

// Package handlers define HTTP handlers.
package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/triage"
	"k8s.io/klog/v2"
)

// Kanban shows a kanban swimlane view of a collection
func (h *Handlers) Kanban() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":          toJS,
		"toYAML":        toYAML,
		"toJSfunc":      toJSfunc,
		"toDays":        toDays,
		"HumanDuration": humanDuration,
		"HumanTime":     humanTime,
		"UnixNano":      unixNano,
		"Avatar":        avatar,
	}
	t := template.Must(template.New("kanban").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "kanban.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		dataAge := time.Time{}
		id := strings.TrimPrefix(r.URL.Path, "/k/")

		defer func() {
			klog.Infof("Served %q request within %s from data %s old", id, time.Since(start), time.Since(dataAge))
		}()

		// 	milestone := getInt(r.URL, "milestone", 0)

		klog.Infof("GET %s (%q): %v", r.URL.Path, id, r.Header)
		s, err := h.party.LookupCollection(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("%q not found: old link or typo?", id), http.StatusNotFound)
			klog.Errorf("collection: %v", err)
			return
		}

		sts, err := h.party.ListCollections()
		if err != nil {
			klog.Errorf("collections: %v", err)
			http.Error(w, "list error", http.StatusInternalServerError)
			return
		}

		var result *triage.CollectionResult
		if isRefresh(r) {
			result = h.updater.ForceRefresh(r.Context(), id)
			klog.Infof("refresh %q result: %d items", id, len(result.RuleResults))
		} else {
			result = h.updater.Lookup(r.Context(), id, true)
			if result == nil {
				http.Error(w, fmt.Sprintf("%q no data", id), http.StatusNotFound)
				return
			}
			if result.RuleResults == nil {
				http.Error(w, fmt.Sprintf("%q no outcomes", id), http.StatusNotFound)
				return
			}

			klog.V(2).Infof("lookup %q result: %d items", id, len(result.RuleResults))
		}

		dataAge = result.Time
		warning := ""

		if time.Since(result.Time) > h.warnAge {
			warning = fmt.Sprintf("Serving results from %s ago. Service started %s ago and is downloading new data. Use Shift-Reload to force refresh at any time.", humanDuration(time.Since(result.Time)), humanDuration(time.Since(h.startTime)))
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
			Types:            "Issues",
			Warning:          warning,
			ResultAge:        time.Since(result.Time),
		}

		for _, s := range sts {
			if s.UsedForStats {
				p.Stats = h.updater.Lookup(r.Context(), s.ID, false)
				p.StatsID = s.ID
			}
		}

		klog.V(2).Infof("page context: %+v", p)
		err = t.ExecuteTemplate(w, "base", p)
		if err != nil {
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}
