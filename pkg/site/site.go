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

	"k8s.io/klog/v2"
)

const VERSION = "v1.1.0"

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
		baseDir:   c.BaseDirectory,
		updater:   c.Updater,
		party:     c.Party,
		siteName:  c.Name,
		warnAge:   c.WarnAge,
		startTime: time.Now(),
	}
}

// Handlers is a mix of config and client interfaces to connect with.
type Handlers struct {
	baseDir   string
	updater   *updater.Updater
	party     *triage.Party
	siteName  string
	warnAge   time.Duration
	startTime time.Time
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
	ResultAge   time.Duration

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

	Swimlanes        []*Swimlane
	CollectionResult *triage.CollectionResult
	SelectorVar      string
	SelectorOptions  []Choice
	Milestone        *github.Milestone

	Stats   *triage.CollectionResult
	StatsID string

	GetVars string
}

// Choice is a selector choice
type Choice struct {
	Value    int
	Text     string
	Selected bool
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

// Make a class name
func className(s string) template.HTMLAttr {
	return template.HTMLAttr(nonWordRe.ReplaceAllString(s, "-"))
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
	os := []*triage.RuleResult{}
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
