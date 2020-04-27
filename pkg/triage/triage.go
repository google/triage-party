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
}

type Party struct {
	engine        *hubbub.Engine
	settings      Settings
	collections   []Collection
	rules         map[string]Rule
	reposOverride []string
}

func New(cfg Config) *Party {
	hc := hubbub.Config{
		Client:      cfg.Client,
		Cache:       cfg.Cache,
		Repos:       cfg.Repos,
		MaxListAge:  cfg.MaxListAge,
		MaxEventAge: cfg.MaxEventAge,
	}
	klog.Infof("New hubbub with config: %+v", hc)
	h := hubbub.New(hc)

	return &Party{
		engine:        h,
		reposOverride: cfg.Repos,
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

	for id, t := range dc.RawRules {
		for _, f := range t.Filters {
			if f.RawLabel != "" {
				err := f.LoadLabelRegex()
				if err != nil {
					return fmt.Errorf("%q: %w", id, err)
				}
			}

			if f.RawTag != "" {
				err := f.LoadTagRegex()
				if err != nil {
					return fmt.Errorf("%q: %w", id, err)
				}

			}
		}
	}

	p.collections = dc.RawCollections
	p.rules = dc.RawRules
	p.settings = dc.Settings
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
