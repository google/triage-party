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

// Package site define HTTP handlers.
package site

import (
	"fmt"
	"html/template"
	"image/color"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/provider"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/triage"
	"github.com/google/triage-party/pkg/updater"

	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v2"

	"k8s.io/klog/v2"
)

// VERSION is what version of Triage Party we advertise as.
const VERSION = "v1.4.0-beta.1"

var (
	nonWordRe = regexp.MustCompile(`\W`)

	// MaxPlayers is how many players to enable in the web interface.
	MaxPlayers = 20

	// Cut-off points for human duration (reversed order)
	defaultMagnitudes = []humanize.RelTimeMagnitude{
		{time.Second, "now", time.Second},
		{2 * time.Second, "1 second %s", 1},
		{time.Minute, "%d seconds %s", time.Second},
		{2 * time.Minute, "1 minute %s", 1},
		{time.Hour, "%d minutes %s", time.Minute},
		{2 * time.Hour, "1 hour %s", 1},
		{humanize.Day, "%d hours %s", time.Hour},
		{2 * humanize.Day, "1 day %s", 1},
		{20 * humanize.Day, "%d days %s", humanize.Day},
		{8 * humanize.Week, "%d weeks %s", humanize.Week},
		{humanize.Year, "%d months %s", humanize.Month},
		{18 * humanize.Month, "1 year %s", 1},
		{2 * humanize.Year, "2 years %s", 1},
		{humanize.LongTime, "%d years %s", humanize.Year},
		{math.MaxInt64, "a long while %s", 1},
	}
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
	Version      string
	SiteName     string
	ID           string
	Title        string
	Description  string
	Warning      template.HTML
	Notification template.HTML
	Total        int
	TotalShown   int
	Types        string
	UniqueItems  []*hubbub.Conversation
	ResultAge    time.Duration
	Stale        bool

	Player        int
	Players       int
	PlayerChoices []string
	PlayerNums    []int
	Index         int

	AverageResponseLatency time.Duration
	TotalPullRequests      int
	TotalIssues            int

	ClosedPerDay float64

	Collection  triage.Collection
	Collections []triage.Collection

	Swimlanes            []*Swimlane
	CollectionResult     *triage.CollectionResult
	SelectorVar          string
	SelectorOptions      []Choice
	Milestone            *provider.Milestone
	CompletionETA        time.Time
	MilestoneETA         time.Time
	MilestoneCountOffset int
	MilestoneVeryLate    bool

	OpenStats     *triage.CollectionResult
	VelocityStats *triage.CollectionResult
	GetVars       string
	Status        string
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
	s = strings.ToLower(nonWordRe.ReplaceAllString(s, "-"))
	s = strings.Replace(s, "_", "-", -1)
	return template.HTMLAttr(s)
}

func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, fmt.Errorf("%q is not a valid hex color", s)
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = fmt.Errorf("%q is not a parseable hex color", s)
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = fmt.Errorf("%q is not a proper hex color", s)
	}
	return
}

// pick an appropriate text color given a background color
func textColor(s string) template.CSS {
	color, err := parseHexColor(fmt.Sprintf("#%s", strings.TrimPrefix(s, "#")))
	if err != nil {
		klog.Errorf("parse hex color failed: %v", err)
		return "f00"
	}

	// human eye is most sensitive to green
	lum := (0.299*float64(color.R) + 0.587*float64(color.G) + 0.114*float64(color.B)) / 255
	if lum > 0.5 {
		return "111"
	}
	return "fff"
}

func unixNano(t time.Time) int64 {
	return t.UnixNano()
}

func humanDuration(d time.Duration) string {
	return roughTime(time.Now().Add(-d))
}

func toDays(d time.Duration) string {
	return fmt.Sprintf("%0.1fd", d.Hours()/24)
}

func roughTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	ds := humanize.CustomRelTime(t, time.Now(), "ago", "from now", defaultMagnitudes)
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

func avatar(u *provider.User) template.HTML {
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
				cs = append(cs, i)
			}
		}

		os = append(os, triage.SummarizeRuleResult(o.Rule, cs, seen))
	}

	return triage.SummarizeCollectionResult(result.Collection, os)
}

// Healthz returns a dummy healthz page - it's always happy here!
func (h *Handlers) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("ok: %s", h.updater.Status())))
	}
}

// Threadz returns a threadz page
func (h *Handlers) Threadz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("GET %s: %v", r.URL.Path, r.Header)
		w.WriteHeader(http.StatusOK)
		w.Write(stack())
	}
}

// stack returns a formatted stack trace of all goroutines
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
