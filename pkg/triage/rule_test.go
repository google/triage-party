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
