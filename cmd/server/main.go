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
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/initcache"
	"github.com/google/triage-party/pkg/site"
	"github.com/google/triage-party/pkg/updater"
)

var (
	configPath    = flag.String("config", "", "configuration path")
	siteDir       = flag.String("site_dir", "site/", "path to site files")
	thirdPartyDir = flag.String("3p_dir", "third_party/", "path to 3rd party files")
	maxListAge    = flag.Duration("max_list_age", 12*time.Hour, "maximum time to cache GitHub searches (prod recommendation: 15s)")
	maxRefreshAge = flag.Duration("max_refresh_age", 15*time.Minute, "Maximum time between strategy runs")
	minRefreshAge = flag.Duration("min_refresh_age", 15*time.Second, "Minimum time between strategy runs")
	warnAge       = flag.Duration("warn_age", 30*time.Minute, "Maximum time before warning about stale results. Recommended: 2*max_refresh_age")

	dryRun    = flag.Bool("dry_run", false, "run queries, don't start a server")
	port      = flag.Int("port", 8080, "port to run server at")
	siteName  = flag.String("site_name", "", "override site name from config file")
	cacheFlag = flag.String("init_cache", "", "Where to load cache from")
	repos     = flag.String("repos", "", "Override configured repos with this repository (comma separated)")
	tokenFlag = flag.String("token", "", "github token")
)

func main() {
	if err := flag.Set("logtostderr", "false"); err != nil {
		panic(fmt.Sprintf("flag set: %v", err))
	}
	if err := flag.Set("alsologtostderr", "true"); err != nil {
		panic(fmt.Sprintf("flag set: %v", err))
	}

	flag.Parse()
	kf := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kf)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := kf.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})

	if *configPath == "" {
		klog.Exitf("--config is required")
	}

	token := os.Getenv("TOKEN")
	if *tokenFlag != "" {
		token = *tokenFlag
	}
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	client := github.NewClient(tc)

	f, err := os.Open(findPath(*configPath))
	if err != nil {
		klog.Exitf("open %s: %v", *configPath, err)
	}

	cachePath := *cacheFlag
	if cachePath == "" {
		name := filepath.Base(*configPath)
		if *repos != "" {
			name = name + "_" + filepath.Base(*repos)
		}
		cachePath = filepath.Join(fmt.Sprintf("/var/tmp/tparty_%s.cache", name))
	}
	klog.Infof("cache path: %s", cachePath)

	c, err := initcache.Load(cachePath)
	if err != nil {
		klog.Exitf("initcache load to %s: %v", cachePath, err)
	}

	cfg := hubbub.Config{
		Client:      client,
		Cache:       c,
		MaxListAge:  *maxListAge,
		MaxEventAge: 90 * 24 * time.Hour,
	}

	if *repos != "" {
		cfg.Repos = strings.Split(*repos, ",")
	}
	h := hubbub.New(cfg)
	if err := h.Load(f); err != nil {
		klog.Exitf("load %s: %v", *configPath, err)
	}

	ts, err := h.ListTactics()
	if err != nil {
		klog.Exitf("list tactics: %v", err)
	}
	klog.Infof("Loaded %d tactics", len(ts))
	sn := *siteName
	if sn == "" {
		sn = calculateSiteName(ts)
	}

	// Make sure save works
	if err := initcache.Save(c, cachePath); err != nil {
		klog.Exitf("initcache save to %s: %v", cachePath, err)
	}

	u := updater.New(updater.Config{
		HubBub:        h,
		Client:        client,
		MinRefreshAge: *minRefreshAge,
		MaxRefreshAge: *maxRefreshAge,
		PersistFunc: func() error {
			return initcache.Save(c, cachePath)
		},
	})

	if *dryRun {
		klog.Infof("Updating ...")
		if err := u.RunOnce(ctx, true); err != nil {
			klog.Exitf("run failed: %v", err)
		}
		os.Exit(0)
	}

	klog.Infof("Starting update loop: %+v", u)
	go u.Loop(ctx)

	s := site.New(&site.Config{
		BaseDirectory: findPath(*siteDir),
		Updater:       u,
		HubBub:        h,
		WarnAge:       *warnAge,
		Name:          sn,
	})

	http.Handle("/third_party/", http.StripPrefix("/third_party/", http.FileServer(http.Dir(findPath(*thirdPartyDir)))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(findPath(*siteDir), "static")))))
	http.HandleFunc("/s/", s.Strategy())
	http.HandleFunc("/", s.Root())

	listenAddr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if listenAddr == ":" {
		listenAddr = fmt.Sprintf(":%d", *port)
	}

	fmt.Printf("\n\n*** teaparty is listening at %s ... ***\n\n", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic(err)
	}
}

// calculates a user-friendly site name based on repositories
func calculateSiteName(ts []hubbub.Tactic) string {
	seen := map[string]bool{}
	for _, t := range ts {
		for _, r := range t.Repos {
			parts := strings.Split(r, "/")
			seen[parts[len(parts)-1]] = true
		}
	}

	names := []string{}
	for n := range seen {
		names = append(names, n)
	}
	return strings.Join(names, " + ")
}

// findPath tries to find the right place for a file
func findPath(p string) string {
	// Running from triage-party/
	if _, err := os.Stat(p); err == nil {
		return p
	}

	// Running from triage-party/cmd/server
	wd, err := os.Getwd()
	if err != nil {
		klog.Errorf("crazy: %v", err)
		return p
	}
	if filepath.Base(wd) == "server" {
		tp := "../../" + p
		if _, err := os.Stat(tp); err == nil {
			return tp
		}
	}
	return p
}
