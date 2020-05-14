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

package triage

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/persist"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

type Config struct {
	Client *github.Client
	Cache  persist.Cacher
	Repos  []string
	// DebugNumber is useful when you want to debug why a single issue is or is-not appearing
	DebugNumber int
}

type Party struct {
	engine        *hubbub.Engine
	settings      Settings
	collections   []Collection
	cache         persist.Cacher
	client        *github.Client
	rules         map[string]Rule
	reposOverride []string
	debugNumber   int
}

func New(cfg Config) *Party {
	p := &Party{
		cache:         cfg.Cache,
		reposOverride: cfg.Repos,
		debugNumber:   cfg.DebugNumber,
		client:        cfg.Client,
	}

	// p.engine is unset until Load() is called
	return p
}

type Settings struct {
	Name          string   `yaml:"name"`
	Repos         []string `yaml:"repos"`
	MinSimilarity float64  `yaml:"min_similarity"`
	MemberRoles   []string `yaml:"member-roles"`
	Members       []string `yaml:"members"`
}

// diskConfig is the on-disk configuration
type diskConfig struct {
	Settings       Settings        `yaml:"settings"`
	RawCollections []Collection    `yaml:"collections"`
	RawRules       map[string]Rule `yaml:"rules"`
}

// newEngine configures a new search engine based on our loaded configs
func (p *Party) newEngine() *hubbub.Engine {
	roles := p.settings.MemberRoles

	if len(roles) == 0 && len(p.settings.Members) == 0 {
		roles = []string{
			"collaborator",
			"member",
			"owner",
		}
	}

	// Why calculate here? So we can share a closed cache among all queries
	maxClosedUpdateAge := time.Duration(0)
	for _, r := range p.rules {
		ca := closedAge(r.Filters)
		if ca > maxClosedUpdateAge {
			maxClosedUpdateAge = ca
		}
	}

	hc := hubbub.Config{
		Client:             p.client,
		Cache:              p.cache,
		Repos:              p.reposOverride,
		DebugNumber:        p.debugNumber,
		MaxClosedUpdateAge: maxClosedUpdateAge,
		MinSimilarity:      p.settings.MinSimilarity,
		MemberRoles:        roles,
		Members:            p.settings.Members,
	}

	klog.Infof("New hubbub with config: %+v", hc)
	return hubbub.New(hc)
}

// Load loads a YAML config from a reader
func (p *Party) Load(r io.Reader) error {

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("readall: %w", err)
	}
	klog.Infof("%d bytes read from config", len(bs))

	dc := &diskConfig{}
	err = yaml.Unmarshal(bs, &dc)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if len(dc.RawCollections) == 0 {
		return fmt.Errorf("no collections found after unmarshal")
	}

	if len(dc.RawRules) == 0 {
		return fmt.Errorf("no rules found after unmarshal")
	}

	rules, err := processRules(dc.RawRules)
	if err != nil {
		return fmt.Errorf("rule processing: %w", err)
	}

	p.collections = dc.RawCollections
	p.rules = rules
	p.settings = dc.Settings

	p.logLoaded()
	if err := p.validateLoadedConfig(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	p.engine = p.newEngine()
	return nil
}

// closedAge returns how old we need to look back for a set of filters
func closedAge(fs []hubbub.Filter) time.Duration {
	oldest := time.Duration(0)
	if !hubbub.NeedsClosed(fs) {
		return oldest
	}

	for _, f := range fs {
		for _, fd := range []string{f.Created, f.Updated, f.Closed, f.Responded} {
			if fd == "" {
				continue
			}

			d, within, _ := hubbub.ParseDuration(fd)
			if !within {
				continue
			}

			if d > oldest {
				oldest = d
			}
		}
	}

	if oldest == 0 {
		klog.Warningf("I need closed data, but I'm not sure how old: picking 4 days")
		return time.Duration(24 * 4 * time.Hour)
	}

	return oldest
}

func (p *Party) validateLoadedConfig() error {
	if len(p.collections) == 0 {
		return fmt.Errorf("no 'collections' defined")
	}
	if len(p.rules) == 0 {
		return fmt.Errorf("no 'rules' defined")
	}

	cols, err := p.ListCollections()
	if err != nil {
		return fmt.Errorf("list collections: %w", err)
	}

	filters := 0
	for _, c := range cols {
		seenRule := map[string]*Rule{}

		for _, tid := range c.RuleIDs {
			if seenRule[tid] != nil {
				return fmt.Errorf("%q has a duplicate rule: %q", c.ID, tid)
			}

			r, err := p.LookupRule(tid)
			if err != nil {
				return fmt.Errorf("lookup rule %q: %w", tid, err)
			}

			seenRule[tid] = &r
			filters += len(r.Filters)
		}
	}

	if filters == 0 {
		return fmt.Errorf("No 'filters' found in the configuration")
	}

	klog.Infof("configuration defines %d filters - looking good!", filters)
	return nil
}

func (p *Party) logLoaded() {
	s, err := yaml.Marshal(p.settings)
	if err != nil {
		klog.Errorf("marshal settings: %v", err)
	}
	klog.Infof("Loaded Settings:\n%s", s)

	s, err = yaml.Marshal(p.collections)
	if err != nil {
		klog.Errorf("marshal collections: %v", err)
	}
	klog.V(2).Infof("Loaded Collections:\n%s", s)

	s, err = yaml.Marshal(p.rules)
	if err != nil {
		klog.Errorf("marshal rules: %v", err)
	}
	klog.V(2).Infof("Loaded Rules:\n%s", s)
}

// processRules precaches regular expressions
func processRules(raw map[string]Rule) (map[string]Rule, error) {
	rules := map[string]Rule{}

	for id, t := range raw {
		rules[id] = t
		newfs := []hubbub.Filter{}

		for _, f := range raw[id].Filters {
			if f.RawLabel != "" {
				err := f.LoadLabelRegex()
				if err != nil {
					return rules, fmt.Errorf("%q label: %w", id, err)
				}
			}

			if f.RawTag != "" {
				err := f.LoadTagRegex()
				if err != nil {
					return rules, fmt.Errorf("%q tag: %w", id, err)
				}
			}

			if f.RawTitle != "" {
				err := f.LoadTitleRegex()
				if err != nil {
					return rules, fmt.Errorf("%q title: %w", id, err)
				}
			}

			newfs = append(newfs, f)
		}

		rules[id] = Rule{
			ID:         t.ID,
			Resolution: t.Resolution,
			Name:       t.Name,
			Repos:      t.Repos,
			Type:       t.Type,
			Filters:    newfs,
		}
	}

	return rules, nil
}
