# Triage Party: Persistent Cache

Triage Party uses an in-memory cache with an optional persistence layer to
significantly speed up startup, as well as decrease load on the GitHub API.

The persistence layer is only read during startup, and is written to
only ocassionally. To configure persistence, use:

* Type: `--persist-backend` flag or `PERSIST_BACKEND` environment variable
* Path: `--persist-path` flag or `PERSIST_PATH` environment flag.

## Disk

Triage Party uses a disk backend by default. It's battle-tested, and ideal for development and smaller deployments. It is not a good match for environments like Google Cloud Run, which do not have persistent storage available.

If `--persist-path` is unset, Triage Party will search for the following directories, choosing the first one which exists.

* `/app/pcache` (production)
* `./pcache`, `../pcache`, `../../pcache` (dev)
* `<UserCacheDir>/pcache` (fallback)

## Google CloudSQL

Triage Party has built-in support for using Google Cloud SQL, using either the MySQL or Postgres backend:

* **MySQL**: `--persist-backend=cloudsql --persist-path="user:password@tcp(project/us-central1/triage-party)/db"`
* **Postgres**: `--persist-backend=cloudsql --persist-path="host=projectname:us-central1:dbname user=postgres password=pw"`

For local development, you will need to setup [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/getting-started).

## MySQL or MariaDB

Example usage:

 `--persist-backend=mysql --persist-path="user:password@tcp(127.0.0.1:3306)/tp"`

## PostgreSQL

Tested with Postgres 11 & 12.2. Example usage:

`--persist-backend=postgres --persist-path="dbname=tp"`

## CockroachDB

CockroachDB has a Postgres front-end, which makes it easy to support. Here's an example, tested with v19.2.6:

 `--persist-backend=postgres postgresql://root@127.0.0.1:26257?sslmode=disable`

## Memory

If no reliable storage is available, this will disable the persistent cache:

`--persist-backend=memory`
