# Triage Party: Deployment Guide

## Environments

### Docker

```shell
docker build --tag=tp --build-arg CFG=examples/generic-project.yaml .
docker run -e GITHUB_TOKEN=<your token> -p 8080:8080 tp
```

### Google Cloud Run

Triage Party was designed to run great with Google Cloud Run. Once a container is built, you can deploy it using the UI, or via gcloud.

As Cloud Run aggressively spins down idle containers and provides no persistant storage, it is highly recommended to persist cache to an external database. Here is an example for persisting to Cloud SQL:

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

## Persistent Storage

Triage Party uses an in-memory cache with an optional persistence layer to decrease the load on GitHub API. By default, Triage Party persists occasionally to disk, but it is configurable via:

* Type: `--persist-backend` flag or `PERSIST_BACKEND` environment variable
* Path: `--persist-path` flag or `PERSIST_PATH` environment flag.

Supported persistence backends include:

* `disk`
* `mem`
* `mysql` (also supports MariaDB)
* `cloudsql` (using MySQL)
* `postgres` (supports PostgreSQL or CockroachDB)

Examples:

* **Custom disk path**: `--persist-path=/var/tmp/tp`
* **MySQL**: `--persist-backend=mysql --persist-path="user:password@tcp(127.0.0.1:3306)/tp"`
* **CloudSQL (MySQL)**: `--persist-backend=cloudsql --persist-path="user:password@tcp(project/us-central1/triage-party)/db"`
  * May require configuring [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/getting-started)
* **PostgreSQL**: `--persist-backend=postgres --persist-path="dbname=tp"`
* **CockroachDB**: `--persist-backend=postgres postgresql://root@127.0.0.1:26257?sslmode=disable`
