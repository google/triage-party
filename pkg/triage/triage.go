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
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

type Config struct {
	Client      *github.Client
	Cache       hubbub.Cacher
	Repos       []string
	MaxListAge  time.Duration
	MaxEventAge time.Duration
	// DebugNumber is useful when you want to debug why a single issue is or is-not appearing
	DebugNumber int
}

type Party struct {
	engine        *hubbub.Engine
	settings      Settings
	collections   []Collection
	rules         map[string]Rule
	reposOverride []string
	debugNumber   int
}

func New(cfg Config) *Party {
	hc := hubbub.Config{
		Client:      cfg.Client,
		Cache:       cfg.Cache,
		Repos:       cfg.Repos,
		MaxListAge:  cfg.MaxListAge,
		MaxEventAge: cfg.MaxEventAge,
		DebugNumber: cfg.DebugNumber,
	}

	klog.Infof("New hubbub with config: %+v", hc)
	h := hubbub.New(hc)

	return &Party{
		engine:        h,
		reposOverride: cfg.Repos,
		debugNumber:   cfg.DebugNumber,
	}
}

type Settings struct {
	Name          string   `yaml:"name"`
	Repos         []string `yaml:"repos"`
	MinSimilarity float64  `yaml:"min_similarity"`
}

// diskConfig is the on-disk configuration
type diskConfig struct {
	Settings       Settings        `yaml:"settings"`
	RawCollections []Collection    `yaml:"collections"`
	RawRules       map[string]Rule `yaml:"rules"`
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

	p.engine.MinSimilarity = dc.Settings.MinSimilarity

	if _, err := p.ListCollections(); err != nil {
		return fmt.Errorf("unable to calculate collections: %v", err)
	}
	p.logLoaded()
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
