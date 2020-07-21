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
	"github.com/google/triage-party/pkg/models"
	"github.com/google/triage-party/pkg/provider"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/persist"
	"github.com/google/triage-party/pkg/triage"

	"k8s.io/klog/v2"
)

var (
	// custom GitHub API URLs
	githubAPIRawURL = flag.String("github-api-url", "", "base URL for GitHub API.  Please set this when you use GitHub Enterprise. This often is your GitHub Enterprise hostname. If the base URL does not have the suffix \"/api/v3/\", it will be added automatically.")

	// shared with server
	configPath      = flag.String("config", "", "configuration path")
	persistBackend  = flag.String("persist-backend", "", "Cache persistence backend (disk, mysql, cloudsql)")
	persistPath     = flag.String("persist-path", "", "Where to persist cache to (automatic)")
	reposOverride   = flag.String("repos", "", "Override configured repos with this repository (comma separated)")
	githubTokenFile = flag.String("github-token-file", "", "github token secret file, also settable via GITHUB_TOKEN")
	numbers         = flag.String("nums", "", "only display results for these comma-delimited issue/PR numbers (debug)")

	// tester specific
	collection = flag.String("collection", "", "collection")
	rule       = flag.String("rule", "", "rule")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if *configPath == "" {
		klog.Exitf("--config is required")
	}

	if *collection == "" && *rule == "" {
		klog.Exitf("--collection or --rule is required")
	}

	ctx := context.Background()

	f, err := os.Open(*configPath)
	if err != nil {
		klog.Exitf("open %s: %v", *configPath, err)
	}

	c, err := persist.FromEnv(*persistBackend, *persistPath, *configPath, *reposOverride)
	if err != nil {
		klog.Exitf("unable to create persistence layer: %v", err)
	}

	if err := c.Initialize(); err != nil {
		klog.Exitf("persist initialize from %s: %v", c, err)
	}

	var debugNums []int
	for _, n := range strings.Split(*numbers, ",") {
		i, err := strconv.Atoi(n)
		if err == nil {
			debugNums = append(debugNums, i)
		}
	}

	initProviderClients(ctx)

	cfg := triage.Config{
		Cache:        c,
		DebugNumbers: debugNums,
	}

	if *reposOverride != "" {
		cfg.Repos = strings.Split(*reposOverride, ",")
	}

	klog.Infof("tester runtime config: %+v", cfg)
	tp := triage.New(cfg)
	if err := tp.Load(f); err != nil {
		klog.Exitf("load %s: %v", *configPath, err)
	}

	if *collection != "" {
		executeCollection(ctx, tp)
	} else {
		executeRule(ctx, tp)
	}
}

// Init providers (Github/Gitlab) HTTP clients
func initProviderClients(ctx context.Context) {
	cfg := provider.Config{
		GithubAPIRawURL: githubAPIRawURL,
		GithubTokenFile: githubTokenFile,
	}
	provider.InitProviders(ctx, cfg)
}

func executeCollection(ctx context.Context, tp *triage.Party) {
	s, err := tp.LookupCollection(*collection)
	if err != nil {
		klog.Exitf("collection: %v", err)
	}

	r, err := tp.ExecuteCollection(ctx, s, time.Now())
	if err != nil {
		klog.Exitf("execute: %v", err)
	}

	fmt.Printf("// Average age: %s\n", toDays(r.AvgAge))
	fmt.Printf("// Average delay: %s\n", toDays(r.AvgAccumulatedHold))
	fmt.Printf("// Average hold: %s\n", toDays(r.AvgCurrentHold))

	for _, o := range r.RuleResults {
		fmt.Printf("## %s\n", o.Rule.Name)
		fmt.Printf(" #  %d items\n", len(o.Items))
		for _, i := range o.Items {
			s, err := json.MarshalIndent(i, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(s))
			fmt.Printf("// Current hold: %s\n", toDays(i.CurrentHoldTime))
			fmt.Printf("// Accumulated hold: %s\n", toDays(i.AccumulatedHoldTime))
		}
	}
}

func executeRule(ctx context.Context, tp *triage.Party) {
	r, err := tp.LookupRule(*rule)
	if err != nil {
		klog.Exitf("rule: %v", err)
	}

	sp := models.SearchParams{
		Ctx:       ctx,
		NewerThan: time.Now(),
		Hidden:    false,
	}

	rr, err := tp.ExecuteRule(sp, r, nil)
	if err != nil {
		klog.Exitf("execute: %v", err)
	}

	fmt.Printf("// Average age: %s\n", toDays(rr.AvgAge))
	fmt.Printf("// Average current hold: %s\n", toDays(rr.AvgCurrentHold))
	fmt.Printf("// Average accumulated hold: %s\n", toDays(rr.AvgAccumulatedHold))

	for _, i := range rr.Items {
		s, err := json.MarshalIndent(i, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(s))
		fmt.Printf("// Current hold: %s\n", toDays(i.CurrentHoldTime))
		fmt.Printf("// Accumulated hold: %s\n", toDays(i.AccumulatedHoldTime))
	}
}

func toDays(d time.Duration) string {
	return fmt.Sprintf("%0.1fd", d.Hours()/24)
}
