#!/bin/bash
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


export PROJECT=k8s-skaffold
export IMAGE=gcr.io/k8s-skaffold/teaparty:`date +%F-%s`
export GITHUB_TOKEN=${GITHUB_TOKEN:-`cat ~/.gh`}
export SERVICE_NAME=skaffold-triage-party

# Resist brew silliness
if [[ -x /usr/bin/python2.7 ]]; then
    export CLOUDSDK_PYTHON=/usr/bin/python2.7
fi

docker build -t $IMAGE \
            --build-arg CFG=examples/skaffold.yaml \
            --build-arg TOKEN=$GITHUB_TOKEN . || exit 1 

docker push $IMAGE || exit 2

gcloud beta run deploy $SERVICE_NAME --project $PROJECT \
                        --image $IMAGE \
                        --set-env-vars=TOKEN=$GITHUB_TOKEN \
                        --allow-unauthenticated \
                        --region us-central1 \
                        --memory 384Mi \
                        --max-instances 1 \
                        --platform managed
