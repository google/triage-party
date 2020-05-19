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

See [deploy/kubernetes](../../deploy/kubernetes) for example manifests. To install Triage Party into a Kubernetes cluster:

```shell
kubectl apply -f deploy/kubernetes
kubectl create secret generic triage-party-github-token -n triage-party --from-file=token=$HOME/.github-token
```

If you are using minikube, this will open Triage Party in your web browser: `minikube service triage-party -n triage-party`

For faster Pod restarts, configure a [persistent cache](persistent.md) using an external database or `PersistentVolumeClaim`

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

For a real-world example deployment script, see [deploy/cloudrun/minikube-deploy.sh](deploy/cloudrun/minikube-deploy.sh)

### Google Cloud Build

```shell
gcloud builds submit .
```

The built image is tagged with `gcr.io/$PROJECT_ID/triage-party:latest`. See the [cloudbuild.yaml](../cloudbuild.yaml) file for more options.
