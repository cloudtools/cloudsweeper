ORG_FILE            := organization.json
WARNING_HOURS		:= 48
DOCKER_GOOGLE_FLAG	:= $(shell echo $${GOOGLE_APPLICATION_CREDENTIALS:+-v ${GOOGLE_APPLICATION_CREDENTIALS}:/google-creds -e GOOGLE_APPLICATION_CREDENTIALS=/google-creds})

build:
	docker build -t cloudsweeper .

clean:
	docker image rm quay.io/agari/cloudsweeper

push: build
	docker push cloudsweeper:latest

run: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper  $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE)

cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) cleanup

reset: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) reset

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) review

mark: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) mark-for-cleanup

warn: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --warning-hours=$(WARNING_HOURS) --org-file=$(ORG_FILE) warn

untagged: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper find-untagged

billing-report: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) billing-report

find: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) --resource-id=$(RESOURCE_ID) find-resource

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-v $(shell pwd)/organization.json:/organization.json \
        -v $(shell pwd)/config.conf:/config.conf \
		--rm -it cloudsweeper setup
