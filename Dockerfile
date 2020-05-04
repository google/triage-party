# syntax = docker/dockerfile:1.0-experimental
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

FROM golang
WORKDIR /app

# CFG is the path to your Triage Party configuration
ARG CFG

ENV SRC_DIR=/src/tparty
ENV GO111MODULE=on

RUN mkdir -p ${SRC_DIR}/cmd ${SRC_DIR}/third_party ${SRC_DIR}/pkg ${SRC_DIR}/site /app/third_party /app/site
COPY go.* $SRC_DIR/
COPY cmd ${SRC_DIR}/cmd/
COPY pkg ${SRC_DIR}/pkg/

# Build the binary
RUN cd $SRC_DIR && go mod download
RUN cd $SRC_DIR/cmd/server && go build -o main
RUN cp $SRC_DIR/cmd/server/main /app/

# Setup our deployment
COPY site /app/site/
COPY third_party /app/third_party/
COPY $CFG /app/config.yaml

# Bad hack: pre-heat the cache in lieu of persistent storage
RUN --mount=type=secret,id=github /app/main --github-token-file=/run/secrets/github --config /app/config.yaml --site /app/site --dry-run

# Run the server at a reasonable refresh rate
CMD ["/app/main", "--min-refresh=25s", "--max-refresh=20m", "--config=/app/config.yaml", "--site=/app/site", "--3p=/app/third_party"]
