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

# This image is used to build the public "triageparty/triage-party"
# image. We hope you enjoy it.
#
# Party on!

FROM golang:latest AS builder
WORKDIR /app

# Build the binary
ENV SRC_DIR=/src/tparty
ENV GO111MODULE=on
RUN mkdir -p ${SRC_DIR}/cmd ${SRC_DIR}/third_party ${SRC_DIR}/pkg ${SRC_DIR}/site /app/third_party /app/site
COPY go.* $SRC_DIR/
COPY cmd ${SRC_DIR}/cmd/
COPY pkg ${SRC_DIR}/pkg/
WORKDIR $SRC_DIR
RUN go mod download
RUN go build cmd/server/main.go

# Setup the site data
# hadolint ignore=DL3007
FROM gcr.io/distroless/base:latest
COPY --from=builder /src/tparty/main /app/
COPY site /app/site/
COPY third_party /app/third_party/
COPY config/config.yaml /app/config/config.yaml

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
