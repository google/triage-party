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

package site

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/triage-party/pkg/persist"
	"github.com/google/triage-party/pkg/triage"
	"github.com/google/triage-party/pkg/updater"
)

type playerParamCase struct {
	name        string
	rawURL      string
	wantPlayers int
	wantPlayer  int
}

func runPlayerParamCase(t *testing.T, tc playerParamCase) {
	t.Helper()
	u, err := url.Parse(tc.rawURL)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", tc.rawURL, err)
	}
	gotPlayers, gotPlayer := parsePlayerParams(u)
	if gotPlayers != tc.wantPlayers || gotPlayer != tc.wantPlayer {
		t.Errorf("parsePlayerParams(%q) = (%d, %d), want (%d, %d)", tc.rawURL, gotPlayers, gotPlayer, tc.wantPlayers, tc.wantPlayer)
	}
}

func TestParsePlayerParams(t *testing.T) {
	cases := []playerParamCase{
		{
			name:        "defaults when omitted",
			rawURL:      "/s/daily",
			wantPlayers: 1,
			wantPlayer:  0,
		},
		{
			name:        "valid multi-player values",
			rawURL:      "/s/daily?players=5&player=3",
			wantPlayers: 5,
			wantPlayer:  3,
		},
		{
			name:        "clamps zero players to minimum",
			rawURL:      "/s/daily?players=0",
			wantPlayers: 1,
			wantPlayer:  0,
		},
		{
			name:        "clamps negative players to minimum",
			rawURL:      "/s/daily?players=-10",
			wantPlayers: 1,
			wantPlayer:  0,
		},
		{
			name:        "clamps huge players to MaxPlayers",
			rawURL:      "/s/daily?players=2000000000",
			wantPlayers: MaxPlayers,
			wantPlayer:  0,
		},
		{
			name:        "resets out-of-range player selection",
			rawURL:      "/s/daily?players=5&player=99",
			wantPlayers: 5,
			wantPlayer:  0,
		},
		{
			name:        "resets negative player selection",
			rawURL:      "/s/daily?players=5&player=-1",
			wantPlayers: 5,
			wantPlayer:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runPlayerParamCase(t, tc)
		})
	}
}

func newTestSiteHandlers(t *testing.T) *Handlers {
	t.Helper()
	cache, err := persist.New(persist.Config{Type: "memory"})
	if err != nil {
		t.Fatalf("persist.New: %v", err)
	}
	if err := cache.Initialize(); err != nil {
		t.Fatalf("cache.Initialize: %v", err)
	}

	tp, err := triage.New(triage.Config{
		Cache:       cache,
		GitHubToken: "dummy-token",
	})
	if err != nil {
		t.Fatalf("triage.New: %v", err)
	}

	const sampleYAML = `settings:
  name: test-party
  repos:
    - https://github.com/google/triage-party
collections:
  - id: home
    name: Home
    rules:
      - sample-rule
rules:
  sample-rule:
    name: Sample Rule
    type: issue
    filters:
      - label: bug
`
	if err := tp.Load(strings.NewReader(sampleYAML)); err != nil {
		t.Fatalf("tp.Load: %v", err)
	}

	u := updater.New(updater.Config{
		Party:      tp,
		MinRefresh: time.Minute,
		MaxRefresh: time.Hour,
	})

	return New(&Config{
		BaseDirectory: filepath.Join("..", "..", "site"),
		Name:          "test-party",
		WarnAge:       time.Hour,
		Updater:       u,
		Party:         tp,
	})
}

func TestCollectionHandlerClampsPlayers(t *testing.T) {
	h := newTestSiteHandlers(t)
	handler := h.Collection()

	req := httptest.NewRequest(http.MethodGet, "/s/home?players=2000000000&player=999", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /s/home?players=2000000000 status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Player 20") {
		t.Errorf("GET /s/home?players=2000000000 response missing clamped Player 20 option")
	}
	if strings.Contains(body, "Player 21") {
		t.Errorf("GET /s/home?players=2000000000 response unexpectedly rendered Player 21 option")
	}
}
