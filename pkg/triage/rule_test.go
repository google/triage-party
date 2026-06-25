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
	"testing"

	"github.com/google/triage-party/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestParseRepo(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    provider.Repo
		wantErr bool
	}{
		{
			name:   "simple github",
			rawURL: "https://github.com/org/repo",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
			},
		},
		{
			name:    "invalid url (no scheme)",
			rawURL:  "github.com/org/repo",
			wantErr: true,
		},
		{
			name:   "gitlab with group",
			rawURL: "https://gitlab.com/org/group/repo",
			want: provider.Repo{
				Host:         "gitlab.com",
				Organization: "org",
				Group:        "group",
				Project:      "repo",
			},
		},
		{
			name:   "single label",
			rawURL: "https://github.com/org/repo?labels=foo",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       []string{"foo"},
			},
		},
		{
			name:   "multiple labels comma separated",
			rawURL: "https://github.com/org/repo?labels=foo,bar",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       []string{"bar", "foo"}, // Should be sorted
			},
		},
		{
			name:   "multiple labels with spaces and empty",
			rawURL: "https://github.com/org/repo?labels= foo , , bar ",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       []string{"bar", "foo"}, // Should be trimmed, cleaned, sorted
			},
		},
		{
			name:   "multiple label params",
			rawURL: "https://github.com/org/repo?labels=foo&labels=bar",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       []string{"bar", "foo"}, // Should be merged and sorted
			},
		},
		{
			name:   "multiple label params with duplicates",
			rawURL: "https://github.com/org/repo?labels=foo,baz&labels=bar,baz",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       []string{"bar", "baz", "foo"}, // Should be merged, de-duplicated, and sorted
			},
		},
		{
			name:   "empty labels param",
			rawURL: "https://github.com/org/repo?labels=",
			want: provider.Repo{
				Host:         "github.com",
				Organization: "org",
				Project:      "repo",
				Labels:       nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := parseRepo(tt.rawURL)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, r)
			}
		})
	}
}
