#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
FROM golang:1.15-alpine

RUN apk update \
 && apk add \
    jq \
    git \
    curl \
    vim \
    librdkafka \
    librdkafka-dev \
    python3 \
 && rm -rf /var/cache/apk/* \
 && ln -s /usr/bin/python3 /usr/bin/python \
 && go get -u github.com/go-delve/delve/cmd/dlv \
 && mkdir /action

WORKDIR /action
ADD proxy /bin/proxy
ADD bin/compile /bin/compile
ADD lib/launcher.go /lib/launcher.go
ENV OW_COMPILER=/bin/compile
ENV OW_LOG_INIT_ERROR=1
ENV OW_WAIT_FOR_ACK=1
ENV OW_EXECUTION_ENV=openwhisk/action-golang-v1.15
ENTRYPOINT [ "/bin/proxy" ]
