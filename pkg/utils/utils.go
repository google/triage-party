package utils

import (
	"fmt"
	"github.com/google/triage-party/pkg/models"
	"net/url"
	"strings"
)

// parseRepo returns provider, organization and project for a URL
// rawURL should be a valid url with host like https://github.com/org/repo
func ParseRepo(rawURL string) (r models.Repo, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if u.Host == "" {
		err = fmt.Errorf("Provided string %s is not a valid URL", rawURL)
		return
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 && len(parts) != 4 {
		// gitlab may have https://gitlab.com/organization/group/repo
		err = fmt.Errorf("expected 2/3 repository parts, got %d: %v", len(parts), parts)
		return
	}
	if len(parts) == 3 {
		r = models.Repo{
			Host:         u.Host,
			Organization: parts[1],
			Project:      parts[2],
		}
	} else {
		r = models.Repo{
			Host:         u.Host,
			Organization: parts[1],
			Group:        parts[2],
			Project:      parts[3],
		}
	}

	return
}
