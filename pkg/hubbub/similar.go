package hubbub

import (
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/imjasonmiller/godice"
	"k8s.io/klog"
)

// A subset of Conversation for related items (requires less memory than a Conversation)
type RelatedConversation struct {
	Organization string `json:"org"`
	Project      string `json:"project"`
	ID           int    `json:"int"`

	URL     string       `json:"url"`
	Title   string       `json:"title"`
	Author  *github.User `json:"author"`
	Type    string       `json:"type"`
	Created time.Time    `json:"created"`
}

func related(c *Conversation) *RelatedConversation {
	return &RelatedConversation{
		Organization: c.Organization,
		Project:      c.Project,
		ID:           c.ID,

		URL:     c.URL,
		Title:   c.Title,
		Author:  c.Author,
		Type:    c.Type,
		Created: c.Created,
	}
}

// updateSimilarConversations updates a slice of conversations with similar ones
func (h *Engine) updateSimilarConversations(cs []*Conversation) error {
	klog.V(2).Infof("updating similar conversations for %d conversations ...", len(cs))
	start := time.Now()
	found := 0
	defer func() {
		klog.V(2).Infof("updated similar conversations for %d conversations within %s", found, time.Since(start))
	}()

	urls, err := h.cachedSimilarURLs()
	if err != nil {
		return err
	}

	for _, c := range cs {
		if len(urls[c.URL]) == 0 {
			continue
		}
		c.Similar = []*RelatedConversation{}
		for _, url := range urls[c.URL] {
			c.Similar = append(c.Similar, related(h.seen[url]))
		}
		found++
	}
	return nil
}

// seenByTitle returns conversations by title
func (h *Engine) seenByTitle() map[string][]*Conversation {
	byTitle := map[string][]*Conversation{}

	for _, c := range h.seen {
		_, ok := byTitle[c.Title]
		if ok {
			byTitle[c.Title] = append(byTitle[c.Title], c)
			continue
		}

		byTitle[c.Title] = []*Conversation{c}
	}
	return byTitle
}

// cachedSimilarURLs returns similar URL's that may be cached
func (h *Engine) cachedSimilarURLs() (map[string][]string, error) {
	if h.similarCacheUpdated.After(h.lastItemUpdate) {
		return h.similarCache, nil
	}

	sim, err := h.similarURLs()
	if err != nil {
		return sim, err
	}

	h.similarCache = sim
	h.similarCacheUpdated = time.Now()
	return h.similarCache, nil
}

// similarURL's returns issue URL's that are similar to one another - SLOW!
func (h *Engine) similarURLs() (map[string][]string, error) {
	if h.MinSimilarity == 0 {
		klog.Warningf("min similarity is 0")
		return nil, nil
	}

	klog.Infof("UPDATING SIMILARITY!")
	start := time.Now()
	defer func() {
		klog.Infof("updateSimilar took %s", time.Since(start))
	}()

	byt := h.seenByTitle()
	titles := []string{}
	for k := range byt {
		titles = append(titles, k)
	}

	sim, err := similarTitles(titles, h.MinSimilarity)
	if err != nil {
		return nil, err
	}

	similar := map[string][]string{}

	for k, v := range sim {
		for _, c := range byt[k] {
			similar[c.URL] = []string{}

			// identical matches
			for _, oc := range byt[k] {
				if oc.URL != c.URL {
					similar[c.URL] = append(similar[c.URL], oc.URL)
				}
			}

			// similar matches
			for _, otherTitle := range v {
				for _, oc := range byt[otherTitle] {
					if oc.URL != c.URL {
						similar[c.URL] = append(similar[c.URL], oc.URL)
					}
				}
			}
		}
	}

	return similar, nil
}

// similarTitles pairs together similar titles - INEFFICIENT
func similarTitles(titles []string, minSimilarity float64) (map[string][]string, error) {
	st := map[string][]string{}

	for _, t := range titles {
		matches, err := godice.CompareStrings(t, titles)
		if err != nil {
			return nil, err
		}

		for _, match := range matches.Candidates {
			if match.Score > minSimilarity {
				if match.Text != t {
					st[t] = append(st[t], match.Text)
				}
			}
		}
	}

	return st, nil
}
