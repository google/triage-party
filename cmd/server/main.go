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

// It's the Triage Party server!
//
// ** Basic example:
//
// go run main.go --github-token-file ~/.token --config minikube.yaml
//
// ** Using MySQL persistence:
//
// --persist-backend=mysql --persist-path="root:rootz@tcp(127.0.0.1:3306)/teaparty"
//

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/triage-party/pkg/constants"

	"k8s.io/klog/v2"

	"github.com/google/triage-party/pkg/persist"
	"github.com/google/triage-party/pkg/provider"
	"github.com/google/triage-party/pkg/site"
	"github.com/google/triage-party/pkg/triage"
	"github.com/google/triage-party/pkg/updater"
)

var (
	// custom GitHub API URLs
	githubAPIRawURL = flag.String("github-api-url", "", "GitHub API url to connect.  Please set this when you use GitHub Enterprise. This often is your GitHub Enterprise hostname. If the URL does not have the suffix \"/api/v3/\", it will be added automatically.")

	// shared with tester
	configPath     = flag.String("config", "", "configuration path (defaults to searching for config.yaml)")
	persistBackend = flag.String("persist-backend", "", "Cache persistence backend (disk, mysql, cloudsql)")
	persistPath    = flag.String("persist-path", "", "Where to persist cache to (automatic)")

	reposOverride   = flag.String("repos", "", "Override configured repos with this repository (comma separated)")
	githubTokenFile = flag.String("github-token-file", "", "github token secret file, also settable via "+constants.GithubTokenEnvVar)
	gitlabTokenFile = flag.String("gitlab-token-file", "", "github token secret file, also settable via "+constants.GitlabTokenEnvVar)

	// server specific
	siteDir       = flag.String("site", "site/", "path to site files")
	thirdPartyDir = flag.String("3p", "third_party/", "path to 3rd party files")
	dryRun        = flag.Bool("dry-run", false, "run queries, don't start a server")
	port          = flag.Int("port", 8080, "port to run server at")
	siteName      = flag.String("name", "", "override site name from config file")
	numbers       = flag.String("nums", "", "only display results for these comma-delimited issue/PR numbers (debug)")

	maxRefresh = flag.Duration("max-refresh", 60*time.Minute, "Maximum time between collection runs")
	minRefresh = flag.Duration("min-refresh", 60*time.Second, "Minimum time between collection runs")
	warnAge    = flag.Duration("warn-age", 90*time.Minute, "Warn when the results are older than this")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	cp := *configPath
	if cp == "" {
		cp = os.Getenv("CONFIG_PATH")
	}
	if cp == "" {
		cp = findPath("config/config.yaml")
		klog.Warningf("--config and CONFIG_PATH were empty, falling back to %s", cp)
	}

	ctx := context.Background()

	f, err := os.Open(findPath(cp))
	if err != nil {
		klog.Exitf("open %s: %v", cp, err)
	}

	c, err := persist.FromEnv(*persistBackend, *persistPath, cp, *reposOverride)
	if err != nil {
		klog.Exitf("unable to create persistence layer: %v", err)
	}

	if err := c.Initialize(); err != nil {
		klog.Exitf("persist initialize for %s: %v", c, err)
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

	klog.Infof("triage runtime config: %+v", cfg)
	tp := triage.New(cfg)
	if err := tp.Load(f); err != nil {
		klog.Exitf("load from %s: %v", cp, err)
	}

	ts, err := tp.ListRules()
	if err != nil {
		klog.Exitf("list rules: %v", err)
	}

	klog.Infof("Loaded %d rules", len(ts))
	sn := *siteName
	if sn == "" {
		sn = calculateSiteName(ts)
	}

	u := updater.New(updater.Config{
		Party:       tp,
		MinRefresh:  *minRefresh,
		MaxRefresh:  *maxRefresh,
		PersistFunc: c.Cleanup,
	})

	if *dryRun {
		klog.Infof("Updating ...")
		if _, err := u.RunOnce(ctx, true); err != nil {
			klog.Exitf("run failed: %v", err)
		}
		os.Exit(0)
	}

	klog.Infof("Starting update loop: %+v", u)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigc
		klog.Infof("signal caught: %v (saving!)", sig)
		u.Persist()
		klog.Infof("Exiting by signal as requested.")
		os.Exit(1)
	}()

	go func() {
		if err := u.Loop(ctx); err == nil {
			klog.Exitf("loop failed: %v", err)
		}
	}()

	s := site.New(&site.Config{
		BaseDirectory: findPath(*siteDir),
		Updater:       u,
		Party:         tp,
		WarnAge:       *warnAge,
		Name:          sn,
	})

	http.Handle("/third_party/", http.StripPrefix("/third_party/", http.FileServer(http.Dir(findPath(*thirdPartyDir)))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(findPath(*siteDir), "static")))))
	http.HandleFunc("/s/", s.Collection())
	http.HandleFunc("/k/", s.Kanban())
	http.HandleFunc("/healthz", s.Healthz())
	http.HandleFunc("/threadz", s.Threadz())

	// In case the previous handlers are removed by errant security systems
	http.HandleFunc("/health", s.Healthz())
	http.HandleFunc("/threads", s.Threadz())

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

// Init providers (Github/Gitlab) HTTP clients
func initProviderClients(ctx context.Context) {
	cfg := provider.Config{
		GithubAPIRawURL: githubAPIRawURL,
		GithubTokenFile: githubTokenFile,
		GitlabTokenFile: gitlabTokenFile,
	}
	provider.InitProviders(ctx, cfg)
}

// calculates a user-friendly site name based on repositories
func calculateSiteName(ts []triage.Rule) string {
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

	prod := filepath.Join("/app/", p)
	if _, err := os.Stat(prod); err == nil {
		return prod
	}

	return p
}
