package hubbub

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/imjasonmiller/godice"
	"k8s.io/klog/v2"
)

var nonLetter = regexp.MustCompile(`[^a-zA-Z]`)
var removeWords = map[string]bool{
	"a":       true,
	"an":      true,
	"and":     true,
	"are":     true,
	"as":      true,
	"be":      true,
	"by":      true,
	"can":     true,
	"does":    true,
	"has":     true,
	"have":    true,
	"how":     true,
	"if":      true,
	"in":      true,
	"is":      true,
	"of":      true,
	"on":      true,
	"or":      true,
	"the":     true,
	"that":    true,
	"to":      true,
	"use":     true,
	"very":    true,
	"via":     true,
	"too":     true,
	"why":     true,
	"add":     true,
	"feature": true,
	"fix":     true,
	"bug":     true,
	"fr":      true,
	"it":      true,
	"you":     true,
	"with":    true,
	"do":      true,
	"we":      true,
}

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

// normalize titles for a higher hit-rate
func normalizeTitle(t string) string {
	keep := []string{}
	for _, word := range strings.Split(t, " ") {
		word = nonLetter.ReplaceAllString(word, "")
		if len(word) == 0 {
			continue
		}
		word = strings.ToLower(word)
		if removeWords[word] {
			continue
		}
		keep = append(keep, word)
	}

	klog.V(3).Infof("normalized: %s", strings.Join(keep, " "))
	return strings.Join(keep, " ")
}

func (h *Engine) updateSimilarityTables(rawTitle, url string) {
	title := normalizeTitle(rawTitle)

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

	// Update us -> them title similarity
	similarTo := []string{}

	h.titleToURLs.Range(func(k, v interface{}) bool {
		otherTitle := k.(string)
		if otherTitle == title {
			return true
		}

		if godice.CompareString(title, otherTitle) > h.MinSimilarity {
			klog.V(1).Infof("%q is similar to %q", rawTitle, otherTitle)
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
	title := normalizeTitle(co.Title)
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
			klog.V(1).Infof("find similar: no conversation found for %s -- must have been filtered out", url)
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
