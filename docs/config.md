# Triage Party: Configuration Guide

## Examples

Triage Party includes with two example configurations that are useful to get started:

* [config](../config/config.yaml): uses label regular expressions that work for most GitHub projects
* [kubernetes](../config/examples/kubernetes.yaml): for projects that use Kubernetes-style labels, particularly prioritization

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
* `closed`: the issue or PR has been closed
* `similar`: the issue or PR appears to be similar to another
* `new-commits`: the PR has new commits since the last member response
