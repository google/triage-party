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

package triage

import (
	"fmt"
	"strings"
	"testing"
)

type configRedactCase struct {
	name       string
	in         Config
	wantGitHub string
	wantGitLab string
}

func runConfigRedactCase(t *testing.T, tc configRedactCase) {
	t.Helper()
	origGitHub := tc.in.GitHubToken
	origGitLab := tc.in.GitLabToken

	got := tc.in.Redacted()
	if got.GitHubToken != tc.wantGitHub {
		t.Errorf("Config.Redacted().GitHubToken = %q, want %q", got.GitHubToken, tc.wantGitHub)
	}
	if got.GitLabToken != tc.wantGitLab {
		t.Errorf("Config.Redacted().GitLabToken = %q, want %q", got.GitLabToken, tc.wantGitLab)
	}
	if tc.in.GitHubToken != origGitHub || tc.in.GitLabToken != origGitLab {
		t.Errorf("Config.Redacted() mutated original tokens: got (%q, %q), want (%q, %q)", tc.in.GitHubToken, tc.in.GitLabToken, origGitHub, origGitLab)
	}
}

func TestConfigRedacted(t *testing.T) {
	cases := []configRedactCase{
		{
			name:       "both tokens set",
			in:         Config{Repos: []string{"google/triage-party"}, GitHubToken: "ghp_secret123", GitLabToken: "glpat_secret456"},
			wantGitHub: "<redacted>",
			wantGitLab: "<redacted>",
		},
		{
			name:       "github set gitlab empty",
			in:         Config{GitHubToken: "ghp_secret123"},
			wantGitHub: "<redacted>",
			wantGitLab: "",
		},
		{
			name:       "both empty",
			in:         Config{},
			wantGitHub: "",
			wantGitLab: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runConfigRedactCase(t, tc)
		})
	}
}

func checkFormattedConfig(t *testing.T, verb string, cfg Config, secret string) {
	t.Helper()
	got := fmt.Sprintf(verb, cfg)
	if strings.Contains(got, secret) {
		t.Errorf("fmt.Sprintf(%q, cfg) = %q, leaked secret %q", verb, got, secret)
	}
	if !strings.Contains(got, "<redacted>") {
		t.Errorf("fmt.Sprintf(%q, cfg) = %q, want <redacted>", verb, got)
	}
}

func TestConfigFormattingRedactsTokens(t *testing.T) {
	secretGH := "ghp_super_secret_token"
	secretGL := "glpat_super_secret_token"
	cfg := Config{
		GitHubAPIURL: "https://api.github.com/",
		GitHubToken:  secretGH,
		GitLabToken:  secretGL,
	}

	for _, verb := range []string{"%v", "%+v", "%s", "%#v"} {
		checkFormattedConfig(t, verb, cfg, secretGH)
		checkFormattedConfig(t, verb, cfg, secretGL)
	}
}
