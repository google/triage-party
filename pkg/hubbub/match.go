package hubbub

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"k8s.io/klog/v2"
)

// Check if an item matches the filters, pre-comment fetch
func preFetchMatch(i GitHubItem, labels []*github.Label, fs []Filter) bool {
	for _, f := range fs {
		klog.V(2).Infof("pre-matching item #%d against filter: %+v", i.GetNumber(), toYAML(f))

		if f.State != "" && f.State != "all" {
			if i.GetState() != f.State {
				klog.V(3).Infof("#%d state is %q, want: %q", i.GetNumber(), i.GetState(), f.State)
				return false
			}
		}

		if f.ClosedCommenters != "" || f.ClosedComments != "" {
			if i.GetState() != "closed" {
				klog.V(3).Infof("#%d state is %q, want closed comments", i.GetNumber(), i.GetState())
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
				klog.V(2).Infof("#%d creation at %s does not meet %s", i.GetNumber(), i.GetCreatedAt(), f.Created)
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

		if f.Milestone != "" {
			if i.GetMilestone().GetTitle() != f.Milestone {
				klog.V(2).Infof("#%d milestone does not meet %s: %+v", i.GetNumber(), f.Milestone, i.GetMilestone())
				return false
			}
		}

		// This state can be performed without downloading comments
		if f.TagRegex() != nil && f.TagRegex().String() == "assigned" {
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
				klog.V(1).Infof("#%d has no updates, but need one for: %s", i.GetNumber(), toYAML(f))
				return false
			}
		}

	}
	return true
}

// Check if an issue matches the summarized version
func postFetchMatch(co *Conversation, fs []Filter) bool {
	for _, f := range fs {
		klog.V(2).Infof("post-matching item #%d against filter: %+v", co.ID, toYAML(f))

		if f.TagRegex() != nil {
			if ok := matchTag(co.Tags, f.TagRegex(), f.TagNegate()); !ok {
				klog.V(4).Infof("#%d did not pass matchTag: %s vs %s %v", co.ID, co.Tags, f.TagRegex(), f.TagNegate())
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

// matchNegateRegex matches a value against a negatable regex
func matchNegateRegex(value string, re *regexp.Regexp, negate bool) bool {
	if re.MatchString(value) {
		klog.V(2).Infof("%q matches %s, returning %v", value, re, !negate)
		return !negate
	}
	// Returns 'false' normally, 'true' when negate is true
	klog.V(2).Infof("%q does not match %s, returning %v", value, re, negate)
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
