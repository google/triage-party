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

// Package site defines HTTP handlers.
package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

// Collection shows a grouping of rules.
func (h *Handlers) Collection() http.HandlerFunc {
	fmap := template.FuncMap{
		"toJS":          toJS,
		"toYAML":        toYAML,
		"toJSfunc":      toJSfunc,
		"toDays":        toDays,
		"HumanDuration": humanDuration,
		"RoughTime":     roughTime,
		"UnixNano":      unixNano,
		"Avatar":        avatar,
		"Class":         className,
		"TextColor":     textColor,
	}
	t := template.Must(template.New("collection").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "collection.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("GET %s: %v", r.URL.Path, r.Header)

		id := strings.TrimPrefix(r.URL.Path, "/s/")
		playerChoices := []string{"Select a player"}
		players := getInt(r.URL, "players", 1)
		player := getInt(r.URL, "player", 0)
		index := getInt(r.URL, "index", 1)

		for i := 0; i < players; i++ {
			playerChoices = append(playerChoices, fmt.Sprintf("Player %d", i+1))
		}

		playerNums := []int{}
		for i := 0; i < MaxPlayers; i++ {
			playerNums = append(playerNums, i+1)
		}

		p, err := h.collectionPage(r.Context(), id, isRefresh(r))
		if err != nil {
			http.Error(w, fmt.Sprintf("collection page for %q: %v", id, err), 500)
			klog.Errorf("page: %v", err)

			return
		}

		result := p.CollectionResult

		if player > 0 && players > 1 {
			p.CollectionResult = playerFilter(result, player, players)
			p.UniqueItems = uniqueItems(p.CollectionResult.RuleResults)
		}

		getVars := ""
		if players > 0 {
			getVars = fmt.Sprintf("?player=%d&players=%d", player, players)
		}

		p.PlayerChoices = playerChoices
		p.PlayerNums = playerNums
		p.Player = player
		p.Players = players
		p.Description = p.Collection.Description
		p.Index = index
		p.GetVars = getVars
		p.ClosedPerDay = calcClosedPerDay(p.VelocityStats)

		err = t.ExecuteTemplate(w, "base", p)

		if err != nil {
			klog.Errorf("tmpl: %v", err)
			return
		}
	}
}
