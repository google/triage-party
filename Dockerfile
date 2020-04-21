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

# CFG is the path to your configuration file
ARG CFG
# TOKEN is your GitHub developer token
ARG TOKEN

# Set an env var that matches your github repo name, replace treeder/dockergo here with your repo name
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

# Setup an initialization cache for warm start-up
RUN env TOKEN=$TOKEN /app/main --config /app/config.yaml --site_dir /app/site --dry_run

# Run the server with tighter cache guarantees
CMD ["/app/main", "--max_list_age=15s", "--max_refresh_age=10m", "--config=/app/config.yaml", "--site_dir=/app/site", "--3p_dir=/app/third_party"]
