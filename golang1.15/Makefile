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
IMG=action-golang-v1.15

build:
	../gradlew distDocker

localbuild:
	GOOS=linux GOARCH=amd64 go build -o proxy -a  -ldflags '-extldflags "-static"' ../main/proxy.go
	docker build -t $(IMG) .
	docker tag $(IMG) $(IMG)

push: build
	docker tag $(IMG) actionloop/$(IMG)
	docker push actionloop/$(IMG):nightly

clean:
	docker rmi -f $(IMG) actionloop/$(IMG)

debug: build
	docker run -p 8080:8080 \
	--name go-action --rm -ti --entrypoint=/bin/bash \
	-e OW_COMPILER=/mnt/bin/compile \
	-v $(PWD):/mnt $(IMG)

enter:
	docker exec -ti go-action bash

run: build
	docker run -d -p 8080:8080 \
	--name go-action --rm \
	-e OW_COMPILER=/mnt/bin/compile \
	-v $(PWD):/mnt $(IMG)

watch:
	docker logs -f go-action

test: run
	$(shell ./test.sh 0 1)
	$(shell ./test.sh 1)
	$(shell ./test.sh 2)

stop:
	docker stop go-action

.PHONY: build push clean debug enter
