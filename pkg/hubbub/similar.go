package hubbub

import (
	"regexp"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/imjasonmiller/godice"
	"k8s.io/klog/v2"
)

var nonLetter = regexp.MustCompile(`[^a-zA-Z]`)

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

func compressTitle(t string) string {
	return nonLetter.ReplaceAllString(t, "")
}

func (h *Engine) updateSimilarityTables(rawTitle, url string) {
	title := compressTitle(rawTitle)

	result, existing := h.titleToURLs.LoadOrStore(title, []string{url})
	if existing {
		foundURL := false
		otherURLs := []string{}
		for _, v := range result.([]string) {
			if v == url {
				foundURL = true
				break
			}
			otherURLs = append(otherURLs, v)
		}

		if !foundURL {
			klog.V(1).Infof("updating %q with %v", rawTitle, otherURLs)
			h.titleToURLs.Store(title, append(otherURLs, url))
		}
		return
	}

	klog.Infof("new title: %q", rawTitle)

	// Update us -> them title similarity
	similarTo := []string{}

	h.titleToURLs.Range(func(k, v interface{}) bool {
		otherTitle := k.(string)
		if otherTitle == title {
			return true
		}

		if godice.CompareString(title, otherTitle) > h.MinSimilarity {
			klog.Infof("%q is similar to %q", rawTitle, otherTitle)
			similarTo = append(similarTo, otherTitle)
		}
		return true
	})

	h.similarTitles.Store(title, similarTo)

	// Update them -> us title similarity
	for _, other := range similarTo {
		klog.V(1).Infof("updating %q to map to %s", other, title)
		others, ok := h.similarTitles.Load(other)
		if ok {
			h.similarTitles.Store(other, append(others.([]string), title))
		}
	}
}

func (h *Engine) FindSimilar(co *Conversation) []*RelatedConversation {
	simco := []*RelatedConversation{}
	title := compressTitle(co.Title)
	similarURLs := []string{}

	tres, ok := h.similarTitles.Load(title)
	if !ok {
		return nil
	}

	for _, ot := range tres.([]string) {
		ures, ok := h.titleToURLs.Load(ot)
		if ok {
			similarURLs = append(similarURLs, ures.([]string)...)
		}
	}

	if len(similarURLs) == 0 {
		return nil
	}

	klog.V(2).Infof("#%d %q is similar to %v", co.ID, co.Title, similarURLs)

	added := map[string]bool{}

	for _, url := range similarURLs {
		// We found ourselves with a different title
		if url == co.URL {
			continue
		}

		// May happen if we've seen a URL with different titles
		if added[url] {
			continue
		}

		oco := h.seen[url]
		if oco == nil {
			klog.Warningf("find similar: no conversation found for %s", url)
			continue
		}
		klog.V(2).Infof("found %s: %q", url, oco.Title)
		simco = append(simco, makeRelated(h.seen[url]))
		added[url] = true
	}
	return simco
}

func makeRelated(c *Conversation) *RelatedConversation {
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
