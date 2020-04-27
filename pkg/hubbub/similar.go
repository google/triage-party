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

// findSimilar returns other conversations related to this URL
func (h *Engine) FindSimilar(c *Conversation) []*RelatedConversation {
	rc := []*RelatedConversation{}

	for _, other := range h.similar[c.URL] {
		oc := h.seen[other]
		rc = append(rc, &RelatedConversation{
			Organization: oc.Organization,
			Project:      oc.Project,
			ID:           oc.ID,

			URL:     oc.URL,
			Title:   oc.Title,
			Author:  oc.Author,
			Type:    oc.Type,
			Created: oc.Created,
		})
	}
	return rc
}

// updateSimilarConversations updates a slice of conversations with similar ones
func (h *Engine) updateSimilarConversations(cs []*Conversation) {
	for _, c := range cs {
		c.Similar = h.FindSimilar(c)
	}
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

// updateSimilar updates the similarity cache
func (h *Engine) updateSimilar() error {

	if h.minSimilarity == 0 {
		return nil
	}

	klog.Infof("UPDATING SIMILARITY!")
	start := time.Now()
	defer func() {
		klog.Infof("updateSimilar took %s", time.Since(start))
	}()

	byt := h.seenByTitle()
	titles := []string{}
	for k, _ := range byt {
		titles = append(titles, k)
	}

	sim, err := similarTitles(titles, h.minSimilarity)
	if err != nil {
		return err
	}

	for k, v := range sim {
		for _, c := range byt[k] {
			h.similar[c.URL] = []string{}

			// identical matches
			for _, oc := range byt[k] {
				if oc.URL != c.URL {
					h.similar[c.URL] = append(h.similar[c.URL], oc.URL)
				}
			}

			// similar matches
			for _, otherTitle := range v {
				for _, oc := range byt[otherTitle] {
					if oc.URL != c.URL {
						h.similar[c.URL] = append(h.similar[c.URL], oc.URL)
					}
				}
			}
		}
	}

	return nil
}
