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
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v24/github"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

// not a regexp
var rawString = regexp.MustCompile(`^[\w-/]+$`)

type Config struct {
	Client      *github.Client
	Cache       Cache
	Repos       []string
	MaxListAge  time.Duration
	MaxEventAge time.Duration
}

type Cache interface {
	Set(string, interface{}, time.Duration)
	Delete(string)
	Get(string) (interface{}, bool)
}

type Settings struct {
	Name          string   `yaml:"name"`
	Repos         []string `yaml:"repos"`
	MinSimilarity float64  `yaml:"min_similarity"`
}

// diskConfig is the on-disk configuration
type diskConfig struct {
	Settings      Settings          `yaml:"settings"`
	RawStrategies []Strategy        `yaml:"strategies"`
	RawTactics    map[string]Tactic `yaml:"tactics"`
}

// Strategy represents a fully loaded YAML configuration
type Strategy struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	TacticIDs    []string `yaml:"tactics"`
	Dedup        bool     `yaml:"dedup,omitempty"`
	Hidden       bool     `yaml:"hidden,omitempty"`
	UsedForStats bool     `yaml:"used_for_statistics,omitempty"`
}

// Tactic is a logical triage group
type Tactic struct {
	ID         string
	Resolution string   `yaml:"resolution,omitempty"`
	Name       string   `yaml:"name,omitempty"`
	Repos      []string `yaml:"repos,omitempty"`
	Type       string   `yaml:"type,omitempty"`
	Filters    []Filter `yaml:"filters"`
}

// Filter lets you do less.
type Filter struct {
	RawLabel    string `yaml:"label,omitempty"`
	labelRegex  *regexp.Regexp
	labelNegate bool

	RawState  string `yaml:"tag,omitempty"`
	tagRegex  *regexp.Regexp
	tagNegate bool

	Milestone string `yaml:"milestone,omitempty"`

	Created            string `yaml:"created,omitempty"`
	Updated            string `yaml:"updated,omitempty"`
	Responded          string `yaml:"responded,omitempty"`
	Reactions          string `yaml:"reactions,omitempty"`
	ReactionsPerMonth  string `yaml:"reactions-per-month,omitempty"`
	Comments           string `yaml:"comments,omitempty"`
	Commenters         string `yaml:"commenters,omitempty"`
	CommentersPerMonth string `yaml:"commenters-per-month,omitempty"`
	ClosedComments     string `yaml:"comments-while-closed,omitempty"`
	ClosedCommenters   string `yaml:"commenters-while-closed,omitempty"`
	State              string `yaml:"state,omitempty"`
}

func (f *Filter) LabelRegex() *regexp.Regexp {
	return f.labelRegex
}

func (f *Filter) LabelNegate() bool {
	return f.labelNegate
}

func (f *Filter) TagRegex() *regexp.Regexp {
	return f.tagRegex
}

func (f *Filter) TagNegate() bool {
	return f.tagNegate
}

func New(cfg Config) *HubBub {
	hb := &HubBub{
		cache:         cfg.Cache,
		client:        cfg.Client,
		maxListAge:    cfg.MaxListAge,
		maxEventAge:   cfg.MaxEventAge,
		reposOverride: cfg.Repos,
		seen:          map[int]*Colloquy{},
		seenTitles:    map[string]int{},
	}
	return hb
}

type HubBub struct {
	cache         Cache
	client        *github.Client
	maxListAge    time.Duration
	maxEventAge   time.Duration
	settings      Settings
	strategies    []Strategy
	seen          map[int]*Colloquy
	seenTitles    map[string]int
	tactics       map[string]Tactic
	reposOverride []string
}

// Return a fully resolved strategy
func (h *HubBub) LookupStrategy(id string) (Strategy, error) {
	for _, s := range h.strategies {
		if s.ID == id {
			return s, nil
		}
	}
	return Strategy{}, fmt.Errorf("%q not found", id)
}

// Return a fully resolved tactic
func (h *HubBub) LookupTactic(id string) (Tactic, error) {
	t, ok := h.tactics[id]
	if !ok {
		return t, fmt.Errorf("tactic %q is undefined - typo?", id)
	}
	t.ID = id
	if len(h.reposOverride) > 0 {
		t.Repos = h.reposOverride
	}

	if len(t.Repos) == 0 {
		t.Repos = h.settings.Repos
	}
	return t, nil
}

// ListStrategies a fully resolved strategies
func (h *HubBub) ListStrategies() ([]Strategy, error) {
	return h.strategies, nil
}

// ListTactics fully resolved tactics
func (h *HubBub) ListTactics() ([]Tactic, error) {
	ts := []Tactic{}
	for k := range h.tactics {
		s, err := h.LookupTactic(k)
		if err != nil {
			return ts, err
		}
		ts = append(ts, s)
	}
	return ts, nil
}

// Flush the search cache for a strategy
func (h *HubBub) FlushSearchCache(id string, minAge time.Duration) error {
	s, err := h.LookupStrategy(id)
	if err != nil {
		return err
	}

	flushed := map[string]bool{}
	for _, tid := range s.TacticIDs {
		t, err := h.LookupTactic(tid)
		if err != nil {
			return err
		}
		for _, r := range t.Repos {
			if !flushed[r] {
				klog.Infof("Flushing search cache for %s ...", r)
				org, project, err := parseRepo(r)
				if err != nil {
					return err
				}
				if err := h.flushIssueSearchCache(org, project, minAge); err != nil {
					klog.Warningf("issue flush for %s/%s: %v", org, project, err)
				}
				if err := h.flushPRSearchCache(org, project, minAge); err != nil {
					klog.Warningf("PR flush for %s/%s: %v", org, project, err)
				}
				flushed[r] = true
			}
		}
	}
	return nil
}

// Load loads a YAML config from a reader
func (h *HubBub) Load(r io.Reader) error {
	dc := &diskConfig{}
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bs, &dc)
	if err != nil {
		return err
	}
	if len(dc.RawStrategies) == 0 {
		return fmt.Errorf("no strategies found")
	}
	if len(dc.RawTactics) == 0 {
		return fmt.Errorf("no tactics found")
	}

	for id, t := range dc.RawTactics {
		for i, f := range t.Filters {
			if f.RawLabel != "" {
				label, negateLabel := negativeMatch(f.RawLabel)
				re, err := regex(label)
				if err != nil {
					return fmt.Errorf("unable to compile regexp for %s label %q: %v", id, label, err)
				}
				t.Filters[i].labelRegex = re
				t.Filters[i].labelNegate = negateLabel
			}

			if f.RawState != "" {
				tag, negateState := negativeMatch(f.RawState)
				re, err := regex(tag)
				if err != nil {
					return fmt.Errorf("unable to compile regexp for %s tag %q: %v", id, tag, err)
				}
				t.Filters[i].tagRegex = re
				t.Filters[i].tagNegate = negateState
			}
		}
	}

	h.strategies = dc.RawStrategies
	h.tactics = dc.RawTactics
	h.settings = dc.Settings
	if _, err := h.ListStrategies(); err != nil {
		return fmt.Errorf("unable to calculate strategies: %v", err)
	}
	h.logLoaded()
	return nil
}

func (h *HubBub) logLoaded() {
	s, err := yaml.Marshal(h.settings)
	if err != nil {
		klog.Errorf("marshal settings: %v", err)
	}
	klog.Infof("Loaded Settings:\n%s", s)

	s, err = yaml.Marshal(h.strategies)
	if err != nil {
		klog.Errorf("marshal strategies: %v", err)
	}
	klog.V(2).Infof("Loaded Strategies:\n%s", s)

	s, err = yaml.Marshal(h.tactics)
	if err != nil {
		klog.Errorf("marshal tactics: %v", err)
	}
	klog.V(2).Infof("Loaded Tactics:\n%s", s)
}

// regex returns regexps matching a string.
func regex(s string) (*regexp.Regexp, error) {
	if rawString.MatchString(s) {
		s = fmt.Sprintf("^%s$", s)
	}
	return regexp.Compile(s)
}

// negativeMatch parses a match string and returns the underlying string and negation bool
func negativeMatch(s string) (string, bool) {
	if strings.HasPrefix(s, "!") {
		return s[1:], true
	}
	return s, false
}
