# Triage Party 🎉

`NOTE: This is not an officially supported Google product`

Triage Party is a tool for triaging incoming GitHub issues for large open-source projects, built with the GitHub API.

![screenshot](screenshot.png)

Triage Party focuses on reducing response latency for incoming GitHub issues and PR's, and ensure that conversations are not lost in the ether. It was built from the [Google Container DevEx team](https://github.com/GoogleContainerTools)'s experience contributing to popular open-source projects, such as [minikube](https://github.com/kubernetes/minikube), [Skaffold](https://github.com/GoogleContainerTools/skaffold/), and [Kaniko](https://github.com/GoogleContainerTools/kaniko/).

Triage Party is a stateless Go web application, configured via YAML. While it has been optimized for Google Cloud Run deployments, it's deployable anywhere due to its low memory footprint: even on a Raspberry Pi.

## Features

* Queries that are not possible on GitHub:
  * conversation state (`tag: recv`, `tag: send`)
  * how long since a project member responded (`responded: +15d`)
  * duration (`updated: +30d`)
  * regexp (`label: priority/.*`)
  * reactions (`reactions: >=5`)
  * comment popularity (`comments-per-month: >0.9`)
  * duplicate detection
  * ... and more!
* Multi-player mode: Supports up to 20 simultaneous players in group triage
* Easily open groups of issues into browser tabs (must allow pop-ups)
* Queries across multiple repositories
* "Shift-Reload" for live data pull

## Triage Party in production

See these fine examples in the wild:

* [kubernetes/minikube](http://tinyurl.com/mk-tparty)
* [GoogleContainerTools/skaffold](http://tinyurl.com/skaffold-tparty)
* [jetstack/cert-manager](https://triage.build-infra.jetstack.net/)

## Requirements

* [GitHub API token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)
* Go v1.14 or higher

## Try it locally

See what Triage Party would look like for an arbitrary repository:

```shell
go run cmd/server/main.go \
  --github-token-file=<path to a file containing your github token> \
  --config examples/generic-kubernetes.yaml \
  --repos kubernetes/sig-release
```

Then visit [http://localhost:8080/](http://localhost:8080/)

The first time you run Triage Party against a new repository, there will be a long delay as it will download data from GitHub. This data will be cached for subsequent runs. We're working to improve this latency.

## Configuration

### Creating a Github token file

1. Create a GitHub token: https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line
2. Store token by pasting it into a text-file:
`echo YOUR_GENERATED_TOKEN > /path/to/file`

### Configuring collections and rules

Each page within Triage Party is represented by a `collection`. Each collection references a list of `rules` that can be shared across collections. Here is a simple collection, which creates a page named `I like soup!`, containing two rules:

```yaml
collections:
  - id: soup
    name: I like soup!
    rules:
      - discuss
      - many-reactions
```

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

For full example configurations, see `examples/*.yaml`. There are two that are particularly useful to get started:

* [generic-project](examples/generic-project.yaml): uses label regular expressions that work for most GitHub projects
* [generic-kubernetes](examples/generic-project.yaml): for projects that use Kubernetes-style labels, particularly  prioritization

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

# Duration since item was created
- created: [-+]duration   # example: +30d
# Duration since item was updated
- updated: [-+]duration
# Duration since item was responded to by a project member
- responded: [-+]duration

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

## Deploying Triage Party

Docker:

```shell
env DOCKER_BUILDKIT=1 \
  GITHUB_TOKEN_PATH=<path to your github token> \
  docker build --tag=tp \
  --build-arg CFG=examples/generic-project.yaml \
  --secret id=github,src=$GITHUB_TOKEN_PATH .

docker run -p 8080:8080 tp
```

Cloud Run:

See [examples/minikube-deploy.sh](examples/minikube-deploy.sh)

Kubernetes:

See [examples/generic-kubernetes.yaml](examples/generic-kubernetes.yaml)
