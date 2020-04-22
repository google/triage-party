# Triage Party 🎉

`NOTE: This is not an officially supported Google product`

Triage Party is a tool for triaging incoming GitHub issues for large open-source projects, built with the GitHub API.

![screenshot](screenshot.png)

Triage focuses on reducing response latency for incoming GitHub issues and PR's, and ensure that conversations are not lost in the ether. It was built from the [Google Container DevEx team](http://github.com/GoogleContainerTools)'s experience contributing to open-source projects, such as minikube, kaniko, and skaffold.

Triage Party is a stateless Go web application, configured via YAML. While it has been optimized for Google Cloud Run deployments, it's deployable anywhere due to it's low memory footprint: even on a Raspberry Pi.

## Features

* Queries across multiple repositories
* Queries that are not possible on GitHub:
  * conversation direction (`tag: recv`)
  * duration (`updated: +30d`)
  * regexp (`label: priority/.*`)
  * reactions (`reactions: >=5`)
  * comment popularity (`comments-per-month: >0.9`)
  * ... and more!
* Multi-player mode: for simultaneous group triage of a pool of issues
* Button to open issue groups as browser tabs (pop-ups must be disabled)
* "Shift-Reload" for live data pull

## Requirements

* [GitHub API token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)
* Go v1.14 or higher

## Triage Party in production

Triage Party is used in production for [kubernetes/minikube](https://github.com/kubernetes/minikube):

* [Triage Party dashboard](http://tinyurl.com/mk-tparty)
* [configuration file](examples/minikube.yaml)
* [Cloud Run deployment script](examples/minikube-deploy.sh)

Other examples:

* [GoogleContainerTools/skaffold Triage Party](https://skaffold-triage-party-eyodkz2nea-uc.a.run.app/)

## Try it locally

See what Triage Party would look like for an arbitrary repository:

```shell
go run cmd/server/main.go \
  --github-token-file=<path to a github token> \
  --config examples/generic-kubernetes.yaml \
  --repos kubernetes/sig-release
```

Then visit [http://localhost:8080/](http://localhost:8080/)

The first time you run Triage Party against a new repository, there will be a long delay as it will download data from GitHub. This data will be cached for subsequent runs. We're working to improve this latency.

## Configuration

Each page is configured with a `strategy` that references multiple queries (`tactics`). These tactics can be shared across pages:

```yaml
strategies:
  - id: soup
    name: I like soup!
    tactics:
      - discuss
      - many-reactions

tactics:
  discuss:
    name: "Items for discussion"
    resolution: "Discuss and remove label"
    filters:
      - label: triage/discuss
      - state: "all"

  many-reactions:
    name: "many reactions, low priority, no recent comment"
    resolution: "Bump the priority, add a comment"
    filters:
      - reactions: ">3"
      - reactions-per-month: ">1"
      - label: "!priority/p0"
      - label: "!priority/p1"
      - responded: +60d
```

For example configurations, see `examples/*.yaml`. There are two that are particularly useful to get started:

* [generic-project](examples/generic-project.yaml): uses label regular expressions that work for most GitHub projects
* [generic-kubernetes](examples/generic-project.yaml): for projects that use Kubernetes-style labels, particularly  prioritization

## Filter language

```yaml
# issue state (default is "open")
- state:(open|closed|all)

# GitHub label
- label: [!]regex

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

## How to deploy

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

See [examples/kubernetes-manifests](examples/kubernetes-manifests)
