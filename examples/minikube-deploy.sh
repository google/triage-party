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

export PROJECT=k8s-minikube
export IMAGE=gcr.io/k8s-minikube/tparty
# You need to set this.
# export GITHUB_TOKEN=
export SERVICE_NAME=teaparty
export CONFIG_FILE=examples/minikube.yaml


docker build -t $IMAGE \
            --build-arg CFG=$CONFIG_FILE \
            --build-arg TOKEN=$GITHUB_TOKEN . || exit 1
docker push $IMAGE || exit 2

gcloud beta run deploy $SERVICE_NAME --project $PROJECT \
                        --image $IMAGE \
                        --set-env-vars=TOKEN=$GITHUB_TOKEN \
                        --allow-unauthenticated \
                        --region us-central1 \
                        --max-instances 2 \
                        --memory 384Mi \
                        --platform managed
 
