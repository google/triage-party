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

export PROJECT=<insert GCP project ID>
export IMAGE=<GCR image, e.g. gcr.io/$PROJECT/myimagename>
export GITHUB_TOKEN=<github access token>
export SERVICE_NAME=<service name>
export CONFIG_FILE=<path to config yaml, e.g. examples/skaffold.yaml>

docker build -t $IMAGE \
            --build-arg CFG=$CONFIG_FILE \
            --build-arg TOKEN=$GITHUB_TOKEN . 
            
docker push $IMAGE

gcloud beta run deploy $SERVICE_NAME --project $PROJECT \
                        --image $IMAGE \
                        --set-env-vars=TOKEN=$GITHUB_TOKEN \
                        --allow-unauthenticated \
                        --region us-central1 \
                        --platform managed

