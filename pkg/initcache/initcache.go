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

// Package initcache provides a bootstrap for the in-memory cache
package initcache

import (
	"time"

	"github.com/google/go-github/v31/github"
)

// Config is cache configuration
type Config struct {
	Type string
	Path string
}

type Hoard struct {
	Creation time.Time
	ID       string

	PullRequests []*github.PullRequest
	Issues       []*github.Issue

	PullRequestComments []*github.PullRequestComment
	IssueComments       []*github.IssueComment

	StringBool map[string]bool
}

// Cacher is the cache interface we support
type Cacher interface {
	Set(string, *Hoard) error
	DeleteOlderThan(string, time.Time) error
	GetNewerThan(string, time.Time) *Hoard

	Initialize() error
	Save() error
}

func New(cfg Config) Cacher {
	if cfg.Type == "disk" {
		return NewDisk(cfg)
	}
	return nil
}
