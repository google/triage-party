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
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"github.com/google/triage-party/pkg/updater"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/v31/github"
	"gopkg.in/yaml.v2"

	"k8s.io/klog"
)

const VERSION = "2020-04-22.01"

var (
	nonWordRe  = regexp.MustCompile(`\W`)
	MaxPlayers = 20
)

// Config is how external users interact with this package.
type Config struct {
	BaseDirectory string
	Name          string
	WarnAge       time.Duration
	Updater       *updater.Updater
	Party         *triage.Party
}

func New(c *Config) *Handlers {
	return &Handlers{
		baseDir:  c.BaseDirectory,
		updater:  c.Updater,
		party:    c.Party,
		siteName: c.Name,
		warnAge:  c.WarnAge,
	}
}

// Handlers is a mix of config and client interfaces to connect with.
type Handlers struct {
	baseDir  string
	updater  *updater.Updater
	party    *triage.Party
	siteName string
	warnAge  time.Duration
}

// Root redirects to leaderboard.
func (h *Handlers) Root() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sts, err := h.party.ListCollections()
		if err != nil {
			klog.Errorf("collections: %v", err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/s/%s", sts[0].ID), http.StatusSeeOther)
	}
}

// Page are values that are passed into the renderer
type Page struct {
	Version     string
	SiteName    string
	ID          string
	Title       string
	Description string
	Warning     string
	Total       int
	TotalShown  int
	Types       string
	UniqueItems []*hubbub.Conversation

	Player        int
	Players       int
	PlayerChoices []string
	PlayerNums    []int
	Mode          int
	Index         int
	EmbedURL      string

	AverageResponseLatency time.Duration
	TotalPullRequests      int
	TotalIssues            int

	Collection  triage.Collection
	Collections []triage.Collection

	CollectionResult *triage.CollectionResult
	Stats            *triage.CollectionResult
	StatsID          string

	GetVars string
}

// is this request an HTTP refresh?
func isRefresh(r *http.Request) bool {
	cc := r.Header["Cache-Control"]
	if len(cc) == 0 {
		return false
	}
	//	klog.Infof("cc=%s headers=%+v", cc, r.Header)
	return cc[0] == "max-age-0" || cc[0] == "no-cache"
}

// Board shows a stratgy board
func (h *Handlers) Collection() http.HandlerFunc {
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
	t := template.Must(template.New("collection").Funcs(fmap).ParseFiles(
		filepath.Join(h.baseDir, "collection.tmpl"),
		filepath.Join(h.baseDir, "base.tmpl"),
	))

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			klog.Infof("Collection request complete in %s", time.Since(start))
		}()
		id := strings.TrimPrefix(r.URL.Path, "/s/")
		playerChoices := []string{"Select a player"}
		players := getInt(r.URL, "players", 1)
		player := getInt(r.URL, "player", 0)
		mode := getInt(r.URL, "mode", 0)
		index := getInt(r.URL, "index", 1)

		for i := 0; i < players; i++ {
			playerChoices = append(playerChoices, fmt.Sprintf("Player %d", i+1))
		}

		playerNums := []int{}
		for i := 0; i < MaxPlayers; i++ {
			playerNums = append(playerNums, i+1)
		}

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

			klog.Infof("lookup %q result: %d items", id, len(result.RuleResults))
		}

		warning := ""
		if time.Since(result.Time) > h.warnAge {
			warning = fmt.Sprintf("Serving stale results (%s old) - refreshing results in background. Use Shift-Reload to force data to refresh at any time.", time.Since(result.Time))
		}

		total := 0
		for _, o := range result.RuleResults {
			total += len(o.Items)
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

		if player > 0 && players > 1 {
			result = playerFilter(result, player, players)
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

		embedURL := ""
		if mode == 1 {
			searchIndex := 0
			for _, o := range result.RuleResults {
				for _, i := range o.Items {
					searchIndex++
					if searchIndex == index {
						embedURL = i.URL
					}
				}
			}
		}

		getVars := ""
		if players > 0 {
			getVars = fmt.Sprintf("?player=%d&players=%d", player, players)
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
			PlayerChoices:    playerChoices,
			PlayerNums:       playerNums,
			Player:           player,
			Players:          players,
			Mode:             mode,
			Index:            index,
			EmbedURL:         embedURL,
			Warning:          warning,
			UniqueItems:      uniqueFiltered,
			GetVars:          getVars,
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

// helper to get integers from a URL
func getInt(url *url.URL, key string, fallback int) int {
	vals := url.Query()[key]
	if len(vals) == 1 {
		i, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			klog.Warningf("bad %s int value: %v", key, vals)
			return fallback
		}
		return int(i)
	}
	return fallback
}

func toYAML(v interface{}) string {
	s, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("yaml err: %v", err)
	}
	return string(s)
}

// Acknowledge JS sanitization issues (what data do we trust?)
func toJS(s string) template.JS {
	return template.JS(s)
}

// Acknowledge JS sanitization issues (what data do we trust?)
func toJSfunc(s string) template.JS {
	return template.JS(nonWordRe.ReplaceAllString(s, "_"))
}

func unixNano(t time.Time) int64 {
	return t.UnixNano()
}

func humanDuration(d time.Duration) string {
	return humanTime(time.Now().Add(-d))
}

func toDays(d time.Duration) string {
	return fmt.Sprintf("%0.1fd", d.Hours()/24)
}

func humanTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	ds := humanize.Time(t)
	ds = strings.Replace(ds, " ago", "", 1)

	ds = strings.Replace(ds, " minutes", "min", 1)
	ds = strings.Replace(ds, " minute", "min", 1)

	ds = strings.Replace(ds, " hours", "h", 1)
	ds = strings.Replace(ds, " hour", "h", 1)

	ds = strings.Replace(ds, " days", "d", 1)
	ds = strings.Replace(ds, " day", "d", 1)

	ds = strings.Replace(ds, " months", "mo", 1)
	ds = strings.Replace(ds, " month", "mo", 1)

	ds = strings.Replace(ds, " years", "y", 1)
	ds = strings.Replace(ds, " year", "y", 1)

	ds = strings.Replace(ds, " weeks", "wk", 1)
	ds = strings.Replace(ds, " week", "wk", 1)

	return ds
}

func avatar(u *github.User) template.HTML {
	return template.HTML(fmt.Sprintf(`<a href="%s" title="%s"><img src="%s" width="20" height="20"></a>`, u.GetHTMLURL(), u.GetLogin(), u.GetAvatarURL()))
}

// playerFilter filters out results for a particular player
func playerFilter(result *triage.CollectionResult, player int, players int) *triage.CollectionResult {
	klog.Infof("Filtering for player %d of %d ...", player, players)
	os := []triage.RuleResult{}

	seen := map[string]*triage.Rule{}

	for _, o := range result.RuleResults {
		cs := []*hubbub.Conversation{}

		for _, i := range o.Items {
			if (i.ID % players) == (player - 1) {
				klog.Infof("%d belongs to player %d", i.ID, player)
				cs = append(cs, i)
			}
		}

		os = append(os, triage.SummarizeRuleResult(o.Rule, cs, seen))

	}
	return triage.SummarizeCollectionResult(os)
}
