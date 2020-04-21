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

// initcache provides a disk cache for getting up and running
package initcache

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/patrickmn/go-cache"
	"golang.org/x/xerrors"
	"k8s.io/klog"

	"github.com/google/triage-party/pkg/hubbub"
)

const (
	ExpireInterval  = 10 * 24 * time.Hour
	CleanupInterval = 15 * time.Minute
)

func init() {
	// Register values we plan to store on disk
	gob.Register(&time.Time{})
	gob.Register(hubbub.IssueCommentCache{})
	gob.Register(hubbub.PRCommentCache{})
	gob.Register(hubbub.IssueSearchCache{})
	gob.Register(hubbub.PRSearchCache{})
	gob.Register(&github.Issue{})
	gob.Register([]hubbub.Colloquy{})
	gob.Register([]*github.Issue{})
	gob.Register([]int{})
	gob.Register(&github.PullRequest{})
	gob.Register([]*github.PullRequest{})
	gob.Register(map[string]bool{})
}

func Create(path string) (*cache.Cache, error) {
	klog.Infof("Creating cache, expire interval: %s", ExpireInterval)
	c := cache.New(ExpireInterval, CleanupInterval)
	c.Set("create-time", time.Now(), ExpireInterval)
	if err := Save(c, path); err != nil {
		return c, xerrors.Errorf("save: %v", err)
	}
	return c, nil
}

func Load(path string) (*cache.Cache, error) {
	klog.Infof("Loading cache from %s ...", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Create(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, xerrors.Errorf("open: %v", err)
	}
	defer f.Close()

	decoded := map[string]cache.Item{}
	d := gob.NewDecoder(bufio.NewReader(f))
	err = d.Decode(&decoded)
	if err != nil && err != io.EOF {
		klog.Errorf("Decode failed: %v", err)
		return Create(path)
	}
	if len(decoded) == 0 {
		return nil, fmt.Errorf("no items loaded from disk: %v", decoded)
	}
	return cache.NewFrom(ExpireInterval, CleanupInterval, decoded), nil
}

func Save(c *cache.Cache, path string) error {
	start := time.Now()
	klog.Infof("Saving items to initcache")
	defer func() {
		klog.Infof("initcache.Save took %s", time.Since(start))
	}()

	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	items := c.Items()
	if err := e.Encode(items); err != nil {
		return xerrors.Errorf("encode: %v", err)
	}
	return ioutil.WriteFile(path, b.Bytes(), 0644)
}
