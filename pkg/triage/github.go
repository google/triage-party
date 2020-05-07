package triage

import (
	"fmt"
	"github.com/google/go-github/v31/github"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

// parseRepo returns the organization and project for a URL
func parseRepo(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(u.Path, "/")

	// not a URL
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	// URL
	if len(parts) == 3 {
		return parts[1], parts[2], nil
	}
	return "", "", fmt.Errorf("expected 2 repository parts, got %d: %v", len(parts), parts)
}

func MustReadToken(path string, env string) string {
	token := os.Getenv(env)
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = strings.TrimSpace(string(t))
		klog.Infof("loaded %d byte github token from %s", len(token), path)
	}

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
