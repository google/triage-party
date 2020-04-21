# Triage Party

Triage Party is a tool for triaging incoming GitHub issues, built with the GitHub API. 

Triage Party was crafted based on our experience triaging issues for Container DevEx OSS projects (minikube, skaffold), with the goal of increasing engineer efficiency and reducing customer response latency. It is a tag-less Go web application, optimized for Google Cloud Run deployments, and designed to be accessible to outside contributors on our projects.

Novel features:
* Shareable bookmarked GitHub queries
* Support for queries across multiple repositories
* Supports queries that are not possible on GitHub:
  * duration (`updated: +30d`)
  * conversation direction (`tag: recv`)
  * regexp (`label: priority/.*`)
  * reactions (`reactions: >=5`)
  * comment popularity (`comments-per-month: >0.9`)
  * ... and more!
* Multi-player mode: for simultaneous group triage of a pool of issues
* Button to open issue group as browser tabs
* High performance through agressive intelligent caching
* Supports "Shift-Reload" for live data pull

Production example: http://tinyurl.com/mk-tparty

![screenshot](screenshot.png)

## Requirements

- [GitHub API token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)
- Go v1.12 or higher

## Checking out the code

```
glogin || prodaccess
git clone git clone sso://user/tstromberg/teaparty
```

## Configuration

See `examples/minikube.yaml`

Supported filters:

```yaml
# issue state (default is "open")
- state:(open|closed|all)

- label: [!]regex
- tag: [!]regex

- milestone: string

- created: [-+]duration
- updated: [-+]duration
- responded: [-+]duration

- reactions: [><=]int
- reactions-per-month: [><=]float

- comments: [><=]int
- comments-per-month: [><=]int
- comments-while-closed: [><=]int

- commenters: [><=]int
- commenters-while-closed: [><=]int
- commenters-per-month: [><=]float
```

## Running locally

Start the webserver:

```
export GO111MODULE=on
cd cmd/server
go run main.go \
  --token $GITHUB_TOKEN \
  --config ../../examples/minikube.yaml
```

This will use minikube's configuration as a starting point. The first time you run Triage Party against a new repository, there will be a long delay as it will download and cache every issue and PR.

## Running in Docker

```
docker build --tag=teaparty --build-arg CFG=examples/minikube.yaml --build-arg TOKEN=$GITHUB_TOKEN .
docker run -p 8080:8080 teaparty
```

## Cloud Run Build & Deploy

An example workflow for Cloud Run builds:

```
export PROJECT=<insert GCP project ID>
export IMAGE=<GCR image, e.g. gcr.io/$PROJECT/myimagename>
export GITHUB_TOKEN=<github access token>
export SERVICE_NAME=<service name>
export CONFIG_FILE=<path to config yaml, e.g. examples/skaffold.yaml>

docker build -t $IMAGE \
            --build-arg CFG=$CONFIG_FILE \
            --build-arg TOKEN=$GITHUB_TOKEN . 
            
docker push $IMAGE

gcloud beta run deploy $SERVICE_NAME --project $PROJECT \
                        --image $IMAGE \
                        --set-env-vars=TOKEN=$GITHUB_TOKEN \
                        --allow-unauthenticated \
                        --region us-central1 \
                        --platform managed
```

Once this project is on GitHub, this button will  work:

[![Run on Google Cloud](https://storage.googleapis.com/cloudrun/button.svg)](https://console.cloud.google.com/cloudshell/editor?shellonly=true&cloudshell_image=gcr.io/cloudrun/button&cloudshell_git_repo=http://github.com/google/triage-party)
