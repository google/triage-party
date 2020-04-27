// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hubbub

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog"
)

var (
	dayRegexp   = regexp.MustCompile(`(\d+)d`)
	weekRegexp  = regexp.MustCompile(`(\d+)w`)
	rangeRegexp = regexp.MustCompile(`([<>=]*)([\d\.]+)`)

	rawString = regexp.MustCompile(`^[\w-/]+$`)
)

// Filter lets you do less.
type Filter struct {
	RawLabel    string `yaml:"label,omitempty"`
	labelRegex  *regexp.Regexp
	labelNegate bool

	RawTag    string `yaml:"tag,omitempty"`
	tagRegex  *regexp.Regexp
	tagNegate bool

	Milestone string `yaml:"milestone,omitempty"`

	Created            string `yaml:"created,omitempty"`
	Updated            string `yaml:"updated,omitempty"`
	Responded          string `yaml:"responded,omitempty"`
	Reactions          string `yaml:"reactions,omitempty"`
	ReactionsPerMonth  string `yaml:"reactions-per-month,omitempty"`
	Comments           string `yaml:"comments,omitempty"`
	Commenters         string `yaml:"commenters,omitempty"`
	CommentersPerMonth string `yaml:"commenters-per-month,omitempty"`
	ClosedComments     string `yaml:"comments-while-closed,omitempty"`
	ClosedCommenters   string `yaml:"commenters-while-closed,omitempty"`
	State              string `yaml:"state,omitempty"`
}

// LoadLabelRegex loads a new label reegx
func (f *Filter) LoadLabelRegex() error {
	label, negateLabel := negativeMatch(f.RawLabel)

	re, err := regex(label)
	if err != nil {
		return err
	}

	f.labelRegex = re
	f.labelNegate = negateLabel
	return nil
}

func (f *Filter) LabelRegex() *regexp.Regexp {
	return f.labelRegex
}

func (f *Filter) LabelNegate() bool {
	return f.labelNegate
}

func (f *Filter) LoadTagRegex() error {
	tag, negateState := negativeMatch(f.RawTag)

	re, err := regex(tag)
	if err != nil {
		return err
	}

	f.tagRegex = re
	f.tagNegate = negateState
	return nil
}

func (f *Filter) TagRegex() *regexp.Regexp {
	return f.tagRegex
}

func (f *Filter) TagNegate() bool {
	return f.tagNegate
}

// Check if an issue matches the summarized version
func matchConversation(co *Conversation, fs []Filter) bool {
	for _, f := range fs {
		if f.TagRegex() != nil {
			if ok := matchTag(co.Tags, f.TagRegex(), f.TagNegate()); !ok {
				klog.V(4).Infof("#%d did not pass matchTag: %s vs %s %s", co.ID, co.Tags, f.TagRegex, f.TagNegate)
				return false
			}
		}
		if f.Responded != "" {
			if ok := matchDuration(co.LatestMemberResponse, f.Responded); !ok {
				klog.V(4).Infof("#%d did not pass matchDuration: %s vs %s", co.ID, co.LatestMemberResponse, f.Responded)
				return false
			}
		}
		if f.Reactions != "" {
			if ok := matchRange(float64(co.ReactionsTotal), f.Reactions); !ok {
				klog.V(2).Infof("#%d did not pass reactions matchRange: %d vs %s", co.ID, co.ReactionsTotal, f.Reactions)
				return false
			}
		}
		if f.Reactions != "" {
			if ok := matchRange(float64(co.ReactionsTotal), f.Reactions); !ok {
				klog.V(2).Infof("#%d did not pass reactions matchRange: %d vs %s", co.ID, co.ReactionsTotal, f.Reactions)
				return false
			}
		}

		if f.ReactionsPerMonth != "" {
			if ok := matchRange(co.ReactionsPerMonth, f.ReactionsPerMonth); !ok {
				klog.V(2).Infof("#%d did not pass reactions per-month matchRange: %f vs %s", co.ID, co.ReactionsPerMonth, f.ReactionsPerMonth)
				return false
			}
		}

		if f.Commenters != "" {
			if ok := matchRange(float64(co.CommentersTotal), f.Commenters); !ok {
				klog.V(2).Infof("#%d did not pass commenters matchRange: %d vs %s", co.ID, co.CommentersTotal, f.Commenters)
				return false
			}
		}

		if f.CommentersPerMonth != "" {
			if ok := matchRange(co.CommentersPerMonth, f.CommentersPerMonth); !ok {
				klog.V(2).Infof("#%d did not pass commenters per-month matchRange: %f vs %s", co.ID, co.CommentersPerMonth, f.CommentersPerMonth)
				return false
			}
		}

		if f.Comments != "" {
			if ok := matchRange(float64(co.CommentsTotal), f.Comments); !ok {
				klog.V(2).Infof("#%d did not pass comments matchRange: %d vs %s", co.ID, co.CommentsTotal, f.Comments)
				return false
			}
		}
		if f.ClosedCommenters != "" {
			if ok := matchRange(float64(co.ClosedCommentersTotal), f.ClosedCommenters); !ok {
				klog.V(2).Infof("#%d did not pass commenters-while-closed matchRange: %d vs %s", co.ID, co.ClosedCommentersTotal, f.ClosedCommenters)
				return false
			}
		}
		if f.ClosedComments != "" {
			if ok := matchRange(float64(co.ClosedCommentsTotal), f.ClosedComments); !ok {
				klog.V(2).Infof("#%d did not pass comments-while-closed matchRange: %d vs %s", co.ID, co.ClosedCommentsTotal, f.ClosedComments)
				return false
			}
		}

	}
	return true
}

func matchLabel(labels []*github.Label, re *regexp.Regexp, negate bool) bool {
	klog.V(2).Infof("Checking label: %s (negate=%v)", re, negate)
	for _, l := range labels {
		if re.MatchString(*l.Name) {
			klog.V(2).Infof("we have a match, returning %v", !negate)
			return !negate
		}
	}
	// Returns 'false' normally, 'true' when negate is true
	klog.V(2).Infof("no match, returning %v", negate)
	return negate
}

func matchTag(tags []string, re *regexp.Regexp, negate bool) bool {
	for _, s := range tags {
		if re.MatchString(s) {
			return !negate
		}
	}
	// Returns 'false' normally, 'true' when negate is true
	return negate
}

func matchDuration(t time.Time, ds string) bool {
	klog.V(2).Infof("match duration: %s vs %s", t, ds)
	// fscking stdlib
	matches := dayRegexp.FindStringSubmatch(ds)
	if len(matches) > 0 {
		d, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			klog.Errorf("unable to parse duration: %s", matches[1])
			return false
		}
		ds = dayRegexp.ReplaceAllString(ds, fmt.Sprintf("%dh", 24*d))
	}

	matches = weekRegexp.FindStringSubmatch(ds)
	if len(matches) > 0 {
		w, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			klog.Errorf("unable to parse duration: %s", matches[1])
			return false
		}
		ds = weekRegexp.ReplaceAllString(ds, fmt.Sprintf("%dh", 24*7*w))
	}

	within := false
	over := false
	if strings.HasPrefix(ds, "-") || strings.HasPrefix(ds, "<") {
		ds = ds[1:]
		within = true
	}

	if strings.HasPrefix(ds, "+") || strings.HasPrefix(ds, ">") {
		ds = ds[1:]
		over = true
	}

	klog.V(2).Infof("parsing duration: %s", ds)
	d, err := time.ParseDuration(ds)
	if err != nil {
		klog.Errorf("unable to parse duration %s: %v", ds, err)
		return false
	}

	if within && time.Since(t) < d {
		return true
	}
	if over && time.Since(t) > d {
		return true
	}
	return false
}

func matchRange(i float64, r string) bool {
	matches := rangeRegexp.FindStringSubmatch(r)
	if len(matches) != 3 {
		klog.Errorf("%q does not match range regexp", r)
		return false
	}

	d, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		klog.Errorf("unable to parse duration: %s", matches[1])
		return false
	}

	switch matches[1] {
	case "":
		return i == d
	case ">":
		return i > d
	case "<":
		return i < d
	case ">=":
		return i >= d
	case "<=":
		return i <= d
	default:
		klog.Errorf("unknown range modifier: %s", matches[1])
		return false
	}
}

// negativeMatch parses a match string and returns the underlying string and negation bool
func negativeMatch(s string) (string, bool) {
	if strings.HasPrefix(s, "!") {
		return s[1:], true
	}
	return s, false
}

// regex returns regexps matching a string.
func regex(s string) (*regexp.Regexp, error) {
	if rawString.MatchString(s) {
		s = fmt.Sprintf("^%s$", s)
	}
	return regexp.Compile(s)
}
