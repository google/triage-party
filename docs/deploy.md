# Triage Party: Deployment Guide

## Environments

### Docker

```shell
docker build --tag=tp --build-arg CFG=examples/generic-project.yaml .
docker run -e GITHUB_TOKEN=<your token> -p 8080:8080 tp
```

### Google Cloud Run

Triage Party was designed to run great with Google Cloud Run. Once a container is built, you can deploy it using the UI, or via gcloud. As Cloud Run aggressively spins down idle containers and provides no persistant storage, it is highly recommended to persist cache to an external database. Here is an example for persisting to Cloud SQL:

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

For mid-to-large repositories, you will want to persist cache to a PersistentVolume or external database in case a pod is rescheduled.

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
* **CloudSQL (MySQL)**: `--persist-backend=cloudsql --persist-path="user:password@tcp(project/us-central1/triage-party)/db"`
  * May require configuring [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/getting-started)
