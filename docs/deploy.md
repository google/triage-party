# Triage Party: Deployment Guide

## Environment variables

While Triage Party primarily uses flags for deployment configuration, several settings are available as environment variables to make it easier to deploy.

* `PORT`: `--port`
* `GITHUB_TOKEN`: (contents of) `--github-token-file`
* `CONFIG_PATH`: `--config`
* `PERSIST_BACKEND`: `--persist-backend`
* `PERSIST_PATH`: `--persist-path`

## Integration

### Docker

The simple Docker deployment is setup for easy cache persistence from disk:

```shell
docker build --tag=tp --build-arg CFG=examples/generic-project.yaml .
docker run -e GITHUB_TOKEN=<your token> -p 8080:8080 tp
```

### Kubernetes

See [examples/manifests](../../examples/manifests)

Add the GitHub token as a secret:

`kubectl create secret generic triage-party-github-token -n triage-party --from-file=token=$HOME/.github-token`

Create a namespace:

`kubectl apply -f examples/manifests/namespace.yaml`

Add the configuration as a ConfigMap, and setup a NodePort:

`kubectl apply -f ./examples/manifests`

For faster Triage Party restarts, configure a [persistent cache](persistent.md).

If you are deploying to minikube, this will open Triage Party up in your local web browser:

`minikube service triage-party -n triage-party`

### Google Cloud Run

Triage Party was designed to run well with Google Cloud Run. Here is an example command-line to deploy against Cloud Run with a Cloud SQL hosted [persistent cache](persist.md).

```shell
gcloud beta run deploy "${SERVICE_NAME}" \
    --project "${PROJECT}" \
    --image "${IMAGE}" \
    --set-env-vars="GITHUB_TOKEN=${token},PERSIST_BACKEND=cloudsql,PERSIST_PATH=tp:${DB_PASS}@tcp(project/region/triage-party)/tp" \
    --allow-unauthenticated \
    --region us-central1 \
    --platform managed
```

For a real-world example deployment script, see [examples/minikube-deploy.sh](examples/minikube-deploy.sh)

### Google Cloud Build

```shell
gcloud builds submit . --substitutions=_CFG=path/to/my/config.yaml
```

The built image is tagged with `gcr.io/$PROJECT_ID/triage-party:latest`. See the [cloudbuild.yaml](../cloudbuild.yaml) file for more options.
