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

package hubbub

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/provider"

	"github.com/imjasonmiller/godice"
	"k8s.io/klog/v2"
)

var (
	nonLetter   = regexp.MustCompile(`[^a-zA-Z]`)
	removeWords = map[string]bool{
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
)

// normalize titles for a higher hit-rate
func normalizeTitle(t string) string {
	var keep []string
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

	klog.V(4).Infof("normalized: %s", strings.Join(keep, " "))
	return strings.Join(keep, " ")
}

// updateSimilarIssues updates similarity tables, meant for background use
func (h *Engine) updateSimilarIssues(key string, is []*provider.Issue) {
	start := time.Now()
	klog.V(1).Infof("Updating similarity table from issue cache %q (%d items)", key, len(is))
	for _, i := range is {
		h.updateSimilarityTables(i.GetTitle(), i.GetHTMLURL())
	}
	klog.V(1).Infof("%q took %s to update", key, time.Since(start))
}

// updateSimilarPullRequests updates similarity tables, meant for background use
func (h *Engine) updateSimilarPullRequests(key string, prs []*provider.PullRequest) {
	start := time.Now()
	klog.V(1).Infof("Updating similarity table from PR cache %q (%d items)", key, len(prs))
	for _, i := range prs {
		h.updateSimilarityTables(i.GetTitle(), i.GetHTMLURL())
	}
	klog.V(1).Infof("%q took %s to update", key, time.Since(start))
}

func (h *Engine) updateSimilarityTables(rawTitle, url string) {
	if h.MinSimilarity == 0 {
		return
	}

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
			klog.V(4).Infof("updating %q with %v", rawTitle, otherURLs)
			h.titleToURLs.Store(title, append(otherURLs, url))
		}
		return
	}

	// Update us -> them title similarity
	similarTo := []string{}

	h.titleToURLs.Range(func(k, v interface{}) bool {
		otherTitle, ok := k.(string)
		if !ok {
			klog.V(1).Infof("key %q is not of type string", k)
		}
		if otherTitle == title {
			return true
		}

		if godice.CompareString(title, otherTitle) > h.MinSimilarity {
			klog.V(4).Infof("%q is similar to %q", rawTitle, otherTitle)
			similarTo = append(similarTo, otherTitle)
		}
		return true
	})

	h.similarTitles.Store(title, similarTo)

	// Update them -> us title similarity
	for _, other := range similarTo {
		klog.V(4).Infof("updating %q to map to %s", other, title)
		others, ok := h.similarTitles.Load(other)
		if ok {
			h.similarTitles.Store(other, append(others.([]string), title))
		}
	}
}

// FindSimilar locates similar conversations to this one
func (h *Engine) FindSimilar(co *Conversation) []*RelatedConversation {
	if h.MinSimilarity == 0 {
		return nil
	}

	simco := []*RelatedConversation{}
	title := normalizeTitle(co.Title)
	similarURLs := []string{}
	klog.V(4).Infof("finding similar items to #%d (%s)", co.ID, co.Type)

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

	klog.V(4).Infof("#%d %q is similar to %v", co.ID, co.Title, similarURLs)

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
			continue
		}

		if oco.Type != co.Type {
			continue
		}

		simco = append(simco, makeRelated(h.seen[url]))
		added[url] = true
	}
	return simco
}
