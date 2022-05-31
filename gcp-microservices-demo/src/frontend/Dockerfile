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

FROM golang:1.17.10-alpine as builder
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates git \
    && apk add build-base \
    && go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /src

# restore dependencies
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .

# build
RUN go build -o /go/bin/frontend .

FROM alpine as release
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates busybox-extras net-tools bind-tools
WORKDIR /src
COPY --from=builder /go/bin/frontend /src/server
COPY ./templates ./templates
COPY ./static ./static

EXPOSE 8080
ENTRYPOINT ["/src/server"]