ORG_FILE            := organization.json
WARNING_HOURS		:= 48
DOCKER_GOOGLE_FLAG	:= $(shell echo $${GOOGLE_APPLICATION_CREDENTIALS:+-v ${GOOGLE_APPLICATION_CREDENTIALS}:/google-creds -e GOOGLE_APPLICATION_CREDENTIALS=/google-creds})

build:
	docker build -t cloudsweeper .

run:
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm cloudsweeper  $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE)

cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) cleanup

reset: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) reset

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-e SMTP_USER \
		-e SMTP_PASS \
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
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --warning-hours=$(WARNING_HOURS) --org-file=$(ORG_FILE) warn

untagged: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm cloudsweeper find-untagged

billing-report: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm cloudsweeper $${CSP:+--csp=${CSP}} --org-file=$(ORG_FILE) billing-report

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		$(DOCKER_GOOGLE_FLAG) \
		--rm -it cloudsweeper setup

test: build
	docker run --rm --entrypoint go cloudsweeper test -cover ./...

build-and-run: build run