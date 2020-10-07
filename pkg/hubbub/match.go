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
	"fmt"
	"github.com/google/triage-party/pkg/provider"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/triage-party/pkg/tag"
	"k8s.io/klog/v2"
)

// Check if an item matches the filters, pre-comment fetch
func preFetchMatch(i provider.IItem, labels []*provider.Label, fs []provider.Filter) bool {
	for _, f := range fs {

		if f.State != "" && f.State != "all" {
			if i.GetState() != f.State {
				return false
			}
		}

		if f.ClosedCommenters != "" || f.ClosedComments != "" {
			if i.GetState() != "closed" {
				return false
			}
		}

		if f.Closed != "" {
			if ok := matchDuration(i.GetClosedAt(), f.Closed); !ok {
				klog.V(2).Infof("#%d closed at %s does not meet %s", i.GetNumber(), i.GetClosedAt(), f.Closed)
				return false
			}
		}

		if f.Updated != "" {
			if ok := matchDuration(i.GetUpdatedAt(), f.Updated); !ok {
				klog.V(2).Infof("#%d update at %s does not meet %s", i.GetNumber(), i.GetUpdatedAt(), f.Updated)
				return false
			}
		}

		if f.Responded != "" {
			if ok := matchDuration(i.GetUpdatedAt(), f.Responded); !ok {
				klog.V(2).Infof("#%d update at %s does not meet responded %s", i.GetNumber(), i.GetUpdatedAt(), f.Responded)
				return false
			}
		}

		if f.Created != "" {
			if ok := matchDuration(i.GetCreatedAt(), f.Created); !ok {
				klog.V(2).Infof("#%d Created at %s does not meet %s", i.GetNumber(), i.GetCreatedAt(), f.Created)
				return false
			}
		}

		if f.TitleRegex() != nil {
			if ok := matchNegateRegex(i.GetTitle(), f.TitleRegex(), f.TitleNegate()); !ok {
				klog.V(2).Infof("#%d title does not meet %s", i.GetNumber(), f.TitleRegex())
				return false
			}
		}

		if f.LabelRegex() != nil {
			if ok := matchLabel(labels, f.LabelRegex(), f.LabelNegate()); !ok {
				klog.V(2).Infof("#%d labels do not meet %s", i.GetNumber(), f.LabelRegex())
				return false
			}
		}

		if f.MilestoneRegex() != nil {
			if ok := matchNegateRegex(i.GetMilestone().GetTitle(), f.MilestoneRegex(), f.MilestoneNegate()); !ok {
				klog.V(2).Infof("#%d milestone does not meet %s", i.GetNumber(), f.MilestoneRegex())
				return false
			}
		}

		// This state can be performed without downloading comments
		if f.TagRegex() != nil && f.TagRegex().String() == "^assigned$" {
			// If assigned and no assignee, fail
			if !f.TagNegate() && i.GetAssignee() == nil {
				return false
			}
			// if !assigned and has assignee, fail
			if f.TagNegate() && i.GetAssignee() != nil {
				return false
			}
		}

		if f.Reactions != "" || f.ReactionsPerMonth != "" || f.Commenters != "" || f.Comments != "" {
			if !i.GetUpdatedAt().After(i.GetCreatedAt()) {
				klog.V(1).Infof("#%d has no updates, but need one for: %s", i.GetNumber(), f)
				return false
			}
		}

	}
	return true
}

// Check if an issue matches the summarized version
func postFetchMatch(co *Conversation, fs []provider.Filter) bool {
	for _, f := range fs {
		klog.V(2).Infof("post-fetch matching item #%d against filter: %+v", co.ID, f)

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

// Check if an issue matches the summarized version, after events have been loaded
func postEventsMatch(co *Conversation, fs []provider.Filter) bool {
	for _, f := range fs {
		if f.TagRegex() != nil {
			if ok, _ := matchTag(co.Tags, f.TagRegex(), f.TagNegate()); !ok {
				klog.V(4).Infof("#%d did not pass matchTag: %s vs %s %v", co.ID, co.Tags, f.TagRegex(), f.TagNegate())
				return false
			}
		}

		if f.Prioritized != "" {
			if ok := matchDuration(co.Prioritized, f.Prioritized); !ok {
				klog.V(4).Infof("#%d did not pass prioritized duration: %s vs %s", co.ID, co.LatestMemberResponse, f.Prioritized)
				return false
			}
		}
	}
	return true
}

func matchLabel(labels []*provider.Label, re *regexp.Regexp, negate bool) bool {
	for _, l := range labels {
		if re.MatchString(*l.Name) {
			return !negate
		}
	}
	// Returns 'false' normally, 'true' when negate is true
	return negate
}

// matchNegateRegex matches a value against a negatable regex
func matchNegateRegex(value string, re *regexp.Regexp, negate bool) bool {
	if value == "" && re.String() != "" && re.String() != "^$" {
		return negate
	}

	if re.MatchString(value) {
		return !negate
	}
	// Returns 'false' normally, 'true' when negate is true
	return negate
}

func matchTag(tags map[tag.Tag]bool, re *regexp.Regexp, negate bool) (bool, tag.Tag) {
	for t := range tags {
		if re.MatchString(t.ID) {
			return !negate, t
		}
	}
	// Returns 'false' normally, 'true' when negate is true
	return negate, tag.None
}

func ParseDuration(ds string) (time.Duration, bool, bool) {
	// fscking stdlib
	matches := dayRegexp.FindStringSubmatch(ds)
	if len(matches) > 0 {
		d, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			klog.Errorf("unable to parse duration: %s", matches[1])
			return 0, false, false
		}
		ds = dayRegexp.ReplaceAllString(ds, fmt.Sprintf("%dh", 24*d))
	}

	matches = weekRegexp.FindStringSubmatch(ds)
	if len(matches) > 0 {
		w, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			klog.Errorf("unable to parse duration: %s", matches[1])
			return 0, false, false
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

	d, err := time.ParseDuration(ds)
	if err != nil {
		klog.Errorf("unable to parse duration %s: %v", ds, err)
		return 0, false, false
	}
	return d, within, over
}

func matchDuration(t time.Time, ds string) bool {
	d, within, over := ParseDuration(ds)

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
