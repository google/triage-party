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

// persist provides a 2-level GitHub cache: in-memory & persistent
package persist

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/google/triage-party/pkg/provider"
)

var inMemoryAge = 7 * 24 * time.Hour

// Config is cache configuration
type Config struct {
	Program string
	Type    string
	Path    string
}

type Blob struct {
	Created time.Time

	// Provider neutral fields, used by triage-party
	PullRequests        []*provider.PullRequest
	Issues              []*provider.Issue
	PullRequestComments []*provider.PullRequestComment
	IssueComments       []*provider.IssueComment
	Timeline            []*provider.Timeline
	Reviews             []*provider.PullRequestReview

	// Provider specific fields, used by other tramps
	GHPullRequest         *github.PullRequest
	GHCommitFiles         []*github.CommitFile
	GHPullRequestComments []*github.PullRequestComment
	GHIssueComments       []*github.IssueComment
	GHIssue               *github.Issue
	GHTimeline            []*github.Timeline
}

// Cacher is the cache interface we support
type Cacher interface {
	String() string

	Set(string, *Blob) error
	Get(string, time.Time) *Blob

	Initialize() error
}

func New(cfg Config) (Cacher, error) {
	gob.Register(&Blob{})
	switch cfg.Type {
	case "mysql":
		return NewMySQL(cfg)
	case "cloudsql":
		return NewCloudSQL(cfg)
	case "postgres":
		return NewPostgres(cfg)
	case "disk", "":
		return NewDisk(cfg)
	case "memory":
		return NewMemory(cfg)
	default:
		return nil, fmt.Errorf("unknown backend: %q", cfg.Type)
	}
}

// FromEnv is shared magic between binaries
func FromEnv(program string, backend string, path string) (Cacher, error) {
	if backend == "" {
		backend = os.Getenv("PERSIST_BACKEND")
	}
	if backend == "" {
		backend = "disk"
	}

	if path == "" {
		path = os.Getenv("PERSIST_PATH")
	}

	if program == "" {
		program = "triage-party"
	}

	c, err := New(Config{
		Program: program,
		Type:    backend,
		Path:    path,
	})

	if err != nil {
		return nil, fmt.Errorf("new from %s: %s: %w", backend, path, err)
	}
	return c, nil
}
