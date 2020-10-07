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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseRepo(t *testing.T) {
	host := "github.com"
	org := "org"
	repo := "repo"
	group := "group"
	u := "https://" + host + "/" + org + "/" + repo
	r, err := parseRepo(u)
	assert.Nil(t, err)
	assert.Equal(t, host, r.Host)
	assert.Equal(t, org, r.Organization)
	assert.Equal(t, repo, r.Project)

	u = host + "/" + org + "/" + repo
	r, err = parseRepo(u)
	assert.NotNil(t, err)
	assert.Equal(t, "", r.Host)
	assert.Equal(t, "", r.Organization)
	assert.Equal(t, "", r.Project)

	u = "https://" + host + "/" + org + "/" + group + "/" + repo
	r, err = parseRepo(u)
	assert.Nil(t, err)
	assert.Equal(t, host, r.Host)
	assert.Equal(t, org, r.Organization)
	assert.Equal(t, repo, r.Project)
	assert.Equal(t, group, r.Group)
}
