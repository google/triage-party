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


# This Dockerfile is optimized for local development or small deployments,
# as it inserts the config file and cached GitHub data into the image.
# Some users may want to create their own Dockerfile or omit it entirely.


# Stage 1: Copy local persistent cache into temp container containing "mv"
FROM alpine AS temp
# CFG is the path to your Triage Party configuration
ARG CFG
COPY pcache /pc
RUN echo "failure is OK with this next step (cache population)"
RUN mv /pc/$(basename $CFG).pc /config.yaml.pc || touch /config.yaml.pc


# Stage 2: Copy persistent cache & configuration into application container
FROM triageparty/triage-party
ARG CFG
COPY --from=temp /config.yaml.pc /app/pcache/config.yaml.pc
COPY site /app/site/
COPY $CFG /app/config/config.yaml


# Useful environment variables:
# 
# * GITHUB_TOKEN: Sets GitHub API token
# * CONFIG_PATH: Sets configuration path (defaults to "/app/config/config.yaml")
# * PORT: Sets HTTP listening port (defaults to 8080)
# 
# For other environment variables, see:
# https://github.com/google/triage-party/blob/master/docs/deploy.md
CMD ["/app/main", "--min-refresh=30s", "--max-refresh=8m", "--site=/app/site", "--3p=/app/third_party"]