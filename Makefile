# Copyright 2021 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

IMAGE_NAME := "rkosegi/owm-exporter"
IMAGE_TAG := "1.0.0"

build-docker:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

build-local:
	go fmt
	go mod download
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o owm-exporter . ; strip owm-exporter

