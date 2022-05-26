#!/bin/bash

# Copyright 2020 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux -o pipefail

git remote -v | grep "google/triage-party.git (fetch)"

if [[ "$(go env GOARCH)" != "amd64" ]]; then
  echo "Please run this on amd64 - buildx has been causing issues"
  exit 1
fi

git fetch
git pull

version=$(grep '^const VERSION' pkg/site/site.go | cut -d'"' -f2)
label=$(echo $version | sed s/^v//g)

echo "version: $version"
echo "label: $label"

docker build -t=triageparty/triage-party -f release.Dockerfile .
docker push triageparty/triage-party:latest

git tag -a $version -m "$version release"
git push --tags

docker tag triageparty/triage-party:latest triageparty/triage-party:$label
docker push triageparty/triage-party:$label
