# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

############################################################################
# About this Dockerfile
#
# This Dockerfile is optimized for local development or deployments which
# require the configuration file to be baked into the resulting image.
#
# If you would rather pass configuration in via other means, such as a
# ConfigMap or environment variable, use the "triageparty/triage-party"
# image published on Docker Hub, or build the equivalent
# using "base.Dockerfile"

# Stage 1: Build Triage Party (identical to base.Dockerfile)
FROM golang:latest AS builder
WORKDIR /app
ENV SRC_DIR=/src/tparty
ENV GO111MODULE=on
RUN mkdir -p ${SRC_DIR}/cmd ${SRC_DIR}/third_party ${SRC_DIR}/pkg ${SRC_DIR}/site /app/third_party /app/site
COPY go.* $SRC_DIR/
COPY cmd ${SRC_DIR}/cmd/
COPY pkg ${SRC_DIR}/pkg/
WORKDIR $SRC_DIR
RUN go mod download
RUN go build cmd/server/main.go

# Stage 2: Copy local persistent cache into temp container containing "mv"
FROM alpine:latest AS temp
ARG CFG=config/config.yaml
COPY pcache /pc
RUN echo "Pre-populating cache if found (failure is perfectly OK)"
RUN mv "/pc/$(basename "${CFG}").pc" /config.yaml.pc || touch /config.yaml.pc

# Stage 3: Build the configured application container
FROM gcr.io/distroless/base:latest AS triage-party
ARG CFG=config/config.yaml
COPY --from=builder /src/tparty/main /app/
COPY --from=temp /config.yaml.pc /app/pcache/config.yaml.pc
COPY site /app/site/
COPY third_party /app/third_party/
COPY $CFG /app/config/config.yaml

# Useful environment variables:
#
# * GITHUB_TOKEN: Sets GitHub API token
# * CONFIG_PATH: Sets configuration path (defaults to "/app/config/config.yaml")
# * PORT: Sets HTTP listening port (defaults to 8080)
# * PERSIST_BACKEND: Set the cache persistence backend
# * PERSIST_PATH: Set the cache persistence path
#
# For other environment variables, see:
# https://github.com/google/triage-party/blob/master/docs/deploy.md
CMD ["/app/main", "--min-refresh=30s", "--max-refresh=8m", "--site=/app/site", "--3p=/app/third_party"]
