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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/initcache"
	"github.com/google/triage-party/pkg/triage"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"
)

var (
	// shared with tester
	configPath      = flag.String("config", "", "configuration path")
	initCachePath   = flag.String("init_cache", "", "Where to load the initial cache from (optional)")
	reposOverride   = flag.String("repos", "", "Override configured repos with this repository (comma separated)")
	githubTokenFile = flag.String("github-token-file", "", "github token secret file, also settable via GITHUB_TOKEN")

	// tester specific
	collection = flag.String("collection", "", "collection")
	number     = flag.Int("num", 0, "only display results for this GitHub number")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	if *configPath == "" {
		klog.Exitf("--config is required")
	}

	if *collection == "" {
		klog.Exitf("--collection is required")
	}

	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: triage.MustReadToken(*githubTokenFile, "GITHUB_TOKEN")},
	)))

	f, err := os.Open(*configPath)
	if err != nil {
		klog.Exitf("open %s: %v", *configPath, err)
	}

	cachePath := *initCachePath
	if cachePath == "" {
		cachePath = initcache.DefaultDiskPath(*configPath, *reposOverride)

	}

	c, err := initcache.Load(cachePath)
	if err != nil {
		klog.Exitf("initcache load to %s: %v", cachePath, err)
	}

	cfg := triage.Config{
		Client:      client,
		Cache:       c,
		MaxListAge:  24 * time.Hour,
		MaxEventAge: 90 * 24 * time.Hour,
	}

	if *reposOverride != "" {
		cfg.Repos = strings.Split(*reposOverride, ",")
	}

	tp := triage.New(cfg)
	if err := tp.Load(f); err != nil {
		klog.Exitf("load %s: %v", *configPath, err)
	}

	s, err := tp.LookupCollection(*collection)
	if err != nil {
		klog.Exitf("collection: %v", err)
	}

	r, err := tp.ExecuteCollection(ctx, client, s)
	if err != nil {
		klog.Exitf("execute: %v", err)
	}
	if err := initcache.Save(c, cachePath); err != nil {
		klog.Exitf("initcache save to %s: %v", cachePath, err)
	}

	for _, o := range r.RuleResults {
		fmt.Printf("## %s\n", o.Rule.Name)
		fmt.Printf(" #  %d items\n", len(o.Items))
		for _, i := range o.Items {
			if *number != 0 && i.ID != *number {
				continue
			}

			s, err := json.MarshalIndent(i, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(s))
			fmt.Printf("// Total Hold: %s\n", i.OnHoldTotal)
			fmt.Printf("// Latest Response Delay: %s\n", i.LatestResponseDelay)
		}
	}
}
