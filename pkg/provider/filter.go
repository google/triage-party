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

package provider

import (
	"fmt"
	"regexp"
	"strings"
)

var rawString = regexp.MustCompile(`^[\w-/]+$`)

// Filter lets you do less.
type Filter struct {
	RawLabel    string `yaml:"label,omitempty"`
	labelRegex  *regexp.Regexp
	labelNegate bool

	RawTag    string `yaml:"tag,omitempty"`
	tagRegex  *regexp.Regexp
	tagNegate bool

	RawTitle    string `yaml:"title,omitempty"`
	titleRegex  *regexp.Regexp
	titleNegate bool

	RawMilestone    string `yaml:"milestone,omitempty"`
	milestoneRegex  *regexp.Regexp
	milestoneNegate bool

	RawAuthor    string `yaml:"author,omitempty"`
	authorRegex  *regexp.Regexp
	authorNegate bool

	Created            string `yaml:"created,omitempty"`
	Updated            string `yaml:"updated,omitempty"`
	Closed             string `yaml:"closed,omitempty"`
	Prioritized        string `yaml:"prioritized,omitempty"`
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

// LoadLabelRegex loads a new label regex
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

// LoadTagRegex loads a new tag regex
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

// LoadTitleRegex loads a new title regex
func (f *Filter) LoadTitleRegex() error {
	r, negateState := negativeMatch(f.RawTitle)

	re, err := regex(r)
	if err != nil {
		return err
	}

	f.titleRegex = re
	f.titleNegate = negateState
	return nil
}

func (f *Filter) TitleRegex() *regexp.Regexp {
	return f.titleRegex
}

func (f *Filter) TitleNegate() bool {
	return f.titleNegate
}

// LoadAuthorRegex loads a new author regex
func (f *Filter) LoadAuthorRegex() error {
	r, negateState := negativeMatch(f.RawAuthor)

	re, err := regex(r)
	if err != nil {
		return err
	}

	f.authorRegex = re
	f.authorNegate = negateState
	return nil
}

func (f *Filter) AuthorRegex() *regexp.Regexp {
	return f.authorRegex
}

func (f *Filter) AuthorNegate() bool {
	return f.authorNegate
}

// LoadMilestoneRegex loads a new milestone regex
func (f *Filter) LoadMilestoneRegex() error {
	r, negateState := negativeMatch(f.RawMilestone)

	re, err := regex(r)
	if err != nil {
		return err
	}

	f.milestoneRegex = re
	f.milestoneNegate = negateState
	return nil
}

func (f *Filter) MilestoneRegex() *regexp.Regexp {
	return f.milestoneRegex
}

func (f *Filter) MilestoneNegate() bool {
	return f.milestoneNegate
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
