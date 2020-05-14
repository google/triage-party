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
	"sync"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/persist"
)

// Config is how to configure a new hubbub engine
type Config struct {
	Client *github.Client // Client is a GitHub client
	Cache  persist.Cacher // Cacher is a cache interface
	Repos  []string       // Repos is the repositories to search

	// MinSimilarity is how close two items need to be to each other to be called similar
	MinSimilarity float64

	// The furthest we will query back for information on closed issues
	MaxClosedUpdateAge time.Duration

	// DebugNumber is used when you want to debug why a single item is being handled in a certain wait
	DebugNumber int
}

// Engine is the search engine interface for hubbub
type Engine struct {
	cache  persist.Cacher
	client *github.Client

	// Must be settable from config
	MinSimilarity float64

	// The furthest we will query back for information on closed issues
	MaxClosedUpdateAge time.Duration

	debugNumber int

	titleToURLs   sync.Map
	similarTitles sync.Map

	// indexes used for similarity matching
	seen map[string]*Conversation
}

func New(cfg Config) *Engine {
	e := &Engine{
		cache:  cfg.Cache,
		client: cfg.Client,

		MaxClosedUpdateAge: cfg.MaxClosedUpdateAge,
		seen:               map[string]*Conversation{},
		MinSimilarity:      cfg.MinSimilarity,
		debugNumber:        cfg.DebugNumber,
	}

	// This value is typically programmed on the fly, but lets give it a good enough default
	if e.MaxClosedUpdateAge == 0 {
		e.MaxClosedUpdateAge = time.Duration(24 * 3 * time.Hour)
	}

	return e
}
