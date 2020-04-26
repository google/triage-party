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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/initcache"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"
)

var (
	tokenFileFlag  = flag.String("github-token-file", "", "github token secret file")
	configFlag     = flag.String("config", "", "configuration path")
	collectionFlag = flag.String("collection", "", "collection")
	cacheFlag      = flag.String("init_cache", "", "Where to load cache from")
	repoFlag       = flag.String("repos", "", "Override configured repos with this repository (comma separated)")
	numFlag        = flag.Int("num", 0, "only display results for this number")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	if *configFlag == "" {
		klog.Exitf("--config is required")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if *tokenFileFlag != "" {
		t, err := ioutil.ReadFile(*tokenFileFlag)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = strings.TrimSpace(string(t))
		klog.Infof("loaded %d byte github token from %s", len(token), *tokenFileFlag)
	}

	if *collectionFlag == "" {
		klog.Exitf("--collection is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	f, err := os.Open(*configFlag)
	if err != nil {
		klog.Exitf("open %s: %v", *configFlag, err)
	}

	cachePath := *cacheFlag
	if cachePath == "" {
		name := filepath.Base(*configFlag)
		if *repoFlag != "" {
			name = name + "_" + filepath.Base(*repoFlag)
		}
		cachePath = filepath.Join(fmt.Sprintf("/var/tmp/tparty_%s.cache", name))
	}

	c, err := initcache.Load(cachePath)
	if err != nil {
		klog.Exitf("initcache load to %s: %v", *cacheFlag, err)
	}

	cfg := hubbub.Config{
		Client:      client,
		Cache:       c,
		MaxListAge:  24 * time.Hour,
		MaxEventAge: 7 * 24 * time.Hour,
	}
	if *repoFlag != "" {
		cfg.Repos = strings.Split(*repoFlag, ",")
	}
	h := hubbub.New(cfg)

	if err := h.Load(f); err != nil {
		klog.Exitf("load: %v", err)
	}

	s, err := h.LookupCollection(*collectionFlag)
	if err != nil {
		klog.Exitf("collection: %v", err)
	}

	r, err := h.ExecuteCollection(ctx, client, s)
	if err != nil {
		klog.Exitf("execute: %v", err)
	}
	if err := initcache.Save(c, *cacheFlag); err != nil {
		klog.Exitf("initcache save to %s: %v", *cacheFlag, err)
	}

	for _, o := range r.RuleResults {
		fmt.Printf("## %s\n", o.Rule.Name)

		for _, i := range o.Items {
			if *numFlag != 0 && i.ID != *numFlag {
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
