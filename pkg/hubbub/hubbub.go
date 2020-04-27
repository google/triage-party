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

// hubbub provides an advanced in-memory search for GitHub using state machines
package hubbub

import (
	"time"

	"github.com/google/go-github/v31/github"
)

// Config is how to configure a new hubbub engine
type Config struct {
	Client      *github.Client // Client is a GitHub client
	Cache       Cacher         // Cacher is a cache interface
	Repos       []string       // Repos is the repositories to search
	MaxListAge  time.Duration  // MaxListAge is when GitHub searches go bad
	MaxEventAge time.Duration  // MaxEventAge is when GitHub events go bad

	// MinSimilarity is how close two items need to be to each other to be called similar
	MinSimilarity float64

	// DebugNumber is used when you want to debug why a single item is being handled in a certain wait
	DebugNumber int
}

// Engine is the search engine interface for hubbub
type Engine struct {
	cache       Cacher
	client      *github.Client
	maxListAge  time.Duration
	maxEventAge time.Duration

	// Must be settable from config
	MinSimilarity float64

	debugNumber int

	// indexes used for similarity matching
	seen map[string]*Conversation
	// stored in-memory only
	similarCache        map[string][]string
	similarCacheUpdated time.Time

	// when did we last see a new item?
	lastItemUpdate time.Time
}

func New(cfg Config) *Engine {
	e := &Engine{
		cache:       cfg.Cache,
		client:      cfg.Client,
		maxListAge:  cfg.MaxListAge,
		maxEventAge: cfg.MaxEventAge,

		seen:          map[string]*Conversation{},
		similarCache:  map[string][]string{},
		MinSimilarity: cfg.MinSimilarity,
		debugNumber:   cfg.DebugNumber,
	}
	return e
}

// Cacher is the supported cache interface for GitHub data
type Cacher interface {
	Set(string, interface{}, time.Duration)
	Delete(string)
	Get(string) (interface{}, bool)
}
