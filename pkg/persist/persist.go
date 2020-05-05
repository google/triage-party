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

// Package persist provides a bootstrap for the in-memory cache
package persist

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v31/github"
)

// Config is cache configuration
type Config struct {
	Type string
	Path string
}

type Thing struct {
	Created time.Time

	PullRequests        []*github.PullRequest
	Issues              []*github.Issue
	PullRequestComments []*github.PullRequestComment
	IssueComments       []*github.IssueComment
	StringBool          map[string]bool
}

// Cacher is the cache interface we support
type Cacher interface {
	String() string

	Set(string, *Thing) error
	DeleteOlderThan(string, time.Time) error
	GetNewerThan(string, time.Time) *Thing

	Initialize() error
	Save() error
}

func New(cfg Config) (Cacher, error) {
	gob.Register(&Thing{})
	switch cfg.Type {
	case "mysql":
		return NewMySQL(cfg)
	case "cloudsql":
		return NewCloudSQL(cfg)
	case "disk", "":
		return NewDisk(cfg)
	case "memory":
		return NewMemory(cfg)
	default:
		return nil, fmt.Errorf("unknown backend: %q", cfg.Type)
	}
}

// FromEnv is shared magic between binaries
func FromEnv(backend string, path string, configPath string, reposOverride string) (Cacher, error) {
	if backend == "" {
		backend = os.Getenv("PERSIST_BACKEND")
	}
	if backend == "" {
		backend = "disk"
	}

	if path == "" {
		path = os.Getenv("PERSIST_PATH")
	}

	if backend == "disk" && path == "" {
		path = DefaultDiskPath(configPath, reposOverride)
	}

	c, err := New(Config{
		Type: backend,
		Path: path,
	})
	if err != nil {
		return nil, fmt.Errorf("new from %s: %s: %w", backend, path, err)
	}
	return c, nil
}
