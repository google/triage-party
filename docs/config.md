# Triage Party: Configuration Guide

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Examples](#examples)
- [Settings](#settings)
- [Collections](#collections)
  - [Settings](#settings-1)
- [Rules](#rules)
- [Filter language](#filter-language)
- [Tags](#tags)
- [Display configuration](#display-configuration)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Examples

Triage Party includes with two example configurations that are useful to get started:

* [config](../config/config.yaml): uses label regular expressions that work for most GitHub projects
* [kubernetes](../config/examples/kubernetes.yaml): for projects that use Kubernetes-style labels, particularly prioritization

## Settings

There are only a handful of site-wide settings worth mentioning:

* `name`: Name of the your Triage Party site
* `min_similarity`: On a scale from 0-1, how similar do two titles need to be before they are labelled as similar. The default is 0 (disabled), but a useful setting is 0.75
* `repos`: A list of repositories to query by default
* `member-roles`: Which GitHub roles to consider as project members
* `members`: A list of people to hard-code as members of the project


## Collections

Each page within Triage Party is represented by a `collection`. Each collection references a list of `rules` that can be shared across collections. Here is a simple collection, which creates a page named `I like soup!`, containing two rules:

```yaml
collections:
  - id: soup
    name: I like soup!
    rules:
      - discuss
      - many-reactions
```

### Settings

For collections, there are a few useful settings to mention:

* `description`: description shown for this collection
* `dedup` (bool): whether to filter out duplicate issues/PR's that show up among multiple rules
* `display`: whether to show this page as `kanban` or `default`
* `overflow`: flag issues if there are issues within a Kanban cell above or equal to this number
* `repos`: an optional list of repos to pull from for this collection
* `category`: an optional category for a hierarchical set of collections

### Scoped Collections

If you maintain a lot of projects, you may want to have a common set of rules applied to different collections, with each collection scoped to a different set of repos.  You may also want to use the category field to create two-level navigation.  For example:

```yaml
# YAML anchor to define a reusable set of rules
triage-rules: &triage-rules
  - issue-needs-type
  - issue-needs-priority
  - question-needs-answer

fix-rules: &fix-rules
  - assigned-issues
  - unassigned-p0-issues
  - unassigned-p1-issues
  - unassigned-p2-issues
  - unassigned-p3-issues
  - unassigned-p4-issues

# YAML anchor to define a reusable set of repos
player-repos: &player-repos
  - https://github.com/shaka-project/shaka-player
  - https://github.com/shaka-project/eme-encryption-scheme-polyfill

# YAML anchor to define a reusable set of repos
packager-repos: &packager-repos
  - https://github.com/shaka-project/shaka-packager

collections:
  - id: shaka-player-triage
    category: Player
    name: Triage
    rules: *triage-rules
    repos: *player-repos

  - id: shaka-player-fix
    category: Player
    name: Fix
    rules: *fix-rules
    repos: *player-repos

  - id: shaka-packager-triage
    category: Packager
    name: Triage
    rules: *triage-rules
    repos: *packager-repos

  - id: shaka-packager-fix
    category: Packager
    name: Fix
    rules: *fix-rules
    repos: *packager-repos
```

## Rules

The first rule, `discuss`, include all items labelled as `triage/discuss`, whether they are pull requests or issues, open or closed.

```yaml
rules:
  discuss:
    name: "Items for discussion"
    resolution: "Discuss and remove label"
    filters:
      - label: triage/discuss
      - state: "all"
```

The second rule, `many-reactions`, is more fine-grained. It is only focused on issues that have seen more than 3 comments, with an average of over 1 reaction per month, is not prioritized highly, and has not seen a response by a member of the project within 2 months:

``` yaml
  many-reactions:
    name: "many reactions, low priority, no recent comment"
    resolution: "Bump the priority, add a comment"
    type: issue
    filters:
      - reactions: ">3"
      - reactions-per-month: ">1"
      - label: "!priority/p0"
      - label: "!priority/p1"
      - responded: +60d
```

## Filter language

```yaml
# issue state (default is "open")
- state:(open|closed|all)

# GitHub label
- label: [!]regex

# Issue or PR title
- title: [!]regex

# Internal tagging: particularly useful tags are:
# - recv: updated by author more recently than a project member
# - recv-q: updated by author with a question
# - send: updated by a project member more recently than the author
- tag: [!]regex

# GitHub milestone
- milestone: string

# Elapsed time since item was created
- created: [-+]duration   # example: +30d
# Elapsed time since item was updated
- updated: [-+]duration
# Elapsed time since item was responded to by a project member
- responded: [-+]duration
# Elapsed time since item was given the current priority
- prioritized: [-+]duration

# Number of reactions this item has received
- reactions: [><=]int  # example: +5
# Number of reactions per month on average
- reactions-per-month: [><=]float

# Number of comments this item has received
- comments: [><=]int
# Number of comments per month on average
- comments-per-month: [><=]int
# Number of comments this item has received while closed!
- comments-while-closed: [><=]int

# Number of commenters on this item
- commenters: [><=]int
# Number of commenters who have interactive with this item while closed
- commenters-while-closed: [><=]int
# Number of commenters tthis item has had per month on average
- commenters-per-month: [><=]float
```

## Tags

Triage Party has an automatic tagging mechanism that adds annotations which can be handy for filtering:

* `commented`: a member of the project has previously commented on this conversation
* `send`: a member of the project added a comment after the author (may be waiting for response from original author)
* `recv`: the original author has commented more recently than a member of the project (may be waiting on a response from a project member)
* `recv-q`: someone asked a question more recently than a member of the project has commented (may be waiting on an answer from a project member)
* `member-last`: a member of the organization was the last commenter
* `author-last`: the original author was the last commenter
* `assigned`: the issue or PR has been assigned to someone
* `assignee-updated`: the issue has been updated by its assignee
* `closed`: the issue or PR has been closed
* `merged`: PR was merged
* `draft`: PR is a draft PR
* `similar`: the issue or PR appears to be similar to another
* `open-milestone`: the issue or PR appears in an open milestone

To determine review state, we support the following tags:

* `approved`: Last review was an approval
* `changes-requested`: Last review was a request for changes
* `reviewed-with-comment`: Last review was a comment
* `new-commits`: the PR has new commits since the last member response
* `unreviewed`: PR has never been reviewed
* `pushed-after-approval`: PR was pushed to after approval

The afforementioned PR review tags are also added to linked issues, though with a `pr-` prefix. For instance, `pr-approved`.

## Display configuration
