package triage

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v31/github"

	"k8s.io/klog/v2"
)

func MustReadToken(path string, env string) string {
	token := os.Getenv(env)
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = string(t)
		klog.Infof("loaded %d byte github token from %s", len(token), path)
	} else {
		klog.Infof("loaded %d byte github token from %s", len(token), env)
	}

	token = strings.TrimSpace(token)
	if len(token) < 8 {
		klog.Exitf("github token impossibly small: %q", token)
	}
	return token
}

func MustCreateGithubClient(githubAPIRawURL string, httpClient *http.Client) *github.Client {
	if githubAPIRawURL != "" {
		client, err := github.NewEnterpriseClient(githubAPIRawURL, githubAPIRawURL, httpClient)
		if err != nil {
			klog.Exitf("unable to create GitHub client: %v", err)
		}
		return client
	}
	return github.NewClient(httpClient)
}
