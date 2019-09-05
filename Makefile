ORG_FILE            	:= organization.json
CONF_FILE           	:= config.conf
WARNING_HOURS		:= 48
DOCKER_GOOGLE_FLAG	:= $(shell echo $${GOOGLE_APPLICATION_CREDENTIALS:+-v ${GOOGLE_APPLICATION_CREDENTIALS}:/google-creds -e GOOGLE_APPLICATION_CREDENTIALS=/google-creds})
CONTAINER_TAG		:= quay.io/agari/cloudsweeper

build:
	docker build -t $(CONTAINER_TAG) .

clean-build:
	docker image rm $(CONTAINER_TAG)

push: build
	docker push $(CONTAINER_TAG):latest

run: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG)

cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) cleanup

reset: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) reset

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) review

mark: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm $(CONTAINER_TAG) mark-for-cleanup

warn: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) warn

untagged: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) find-untagged

billing-report: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) billing-report

find: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm $(CONTAINER_TAG) --resource-id=$(RESOURCE_ID) find-resource

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/$(ORG_FILE):/$(ORG_FILE) \
		-v $(shell pwd)/$(CONF_FILE):/$(CONF_FILE) \
		--rm -it $(CONTAINER_TAG) setup
