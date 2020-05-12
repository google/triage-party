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

```shell
docker build --tag=tp --build-arg CFG=examples/generic-project.yaml .
docker run -e GITHUB_TOKEN=<your token> -p 8080:8080 tp
```

### Google Cloud Build

```shell
gcloud builds submit . --substitutions=_CFG=path/to/my/config.yaml
```

The built image is tagged with `gcr.io/$PROJECT_ID/triage-party:latest`. See the [cloudbuild.yaml](../cloudbuild.yaml) file for more options.

### Google Cloud Run

Triage Party was designed to run with Google Cloud Run. That said, Google Cloud Run agressively spins down containers which are not serving incoming requests, which means:

* Results may sometimes be stale (a warning will be shown in the UI)
* For faster start-up, you'll want to configure an external persistent cache, such as Cloud SQL

Here is a command-line example that deploys to Cloud Run with Cloud SQL persistence:

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

### Kubernetes

See [examples/generic-kubernetes.yaml](examples/generic-kubernetes.yaml)

For faster start-up, you will want to persist cache externally to a PersistentVolume or database in case a pod is rescheduled.

## Configuring Persistence

Triage Party uses an in-memory cache with an optional persistence layer to decrease the load on GitHub API. It uses disk by default, but can be configured to use an external databasee. To configure it, use:

* Backend type: `--persist-backend` flag or `PERSIST_BACKEND` environment variable
* Backend path: `--persist-path` flag or `PERSIST_PATH` environment flag.

Supported backends include:

* `disk` - useful for development or small installations
* `mysql` - useful for all installations
* `postgres` - supports both PostgreSQL and CockroachDB
* `cloudsql` - useful for Google Cloud installations
* `memory` - no persistence

Examples flag settings:

* **Custom disk path**: `--persist-path=/var/tmp/tp`
* **MySQL**: `--persist-backend=mysql --persist-path="user:password@tcp(127.0.0.1:3306)/tp"`
* **PostgreSQL**: `--persist-backend=postgres --persist-path="dbname=tp"`
* **CockroachDB**: `--persist-backend=postgres postgresql://root@127.0.0.1:26257?sslmode=disable`
* **CloudSQL - MySQL**: `--persist-backend=cloudsql --persist-path="user:password@tcp(project/us-central1/triage-party)/db"`
* **CloudSQL - Postgres**: `--persist-backend=cloudsql --persist-path="host=projectname:us-central1:dbname user=postgres password=pw"`

NOTE: Local development with CloudSQL backends may require setting up [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/getting-started)
