ALL_ACCOUNTS 		:= aws_accounts.json
FRIENDLIES 			:= friendly_accounts.json
WARNING_HOURS		:= 48

build:
	docker build -t housekeeper .

run:
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --accounts-file=$(FRIENDLIES)

cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --accounts-file=$(FRIENDLIES) cleanup

reset: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --accounts-file=$(FRIENDLIES) reset

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --accounts-file=$(FRIENDLIES) review

mark: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --accounts-file=$(FRIENDLIES) mark-for-cleanup

warn: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --warning-hours=$(WARNING_HOURS) --accounts-file=$(FRIENDLIES) warn

billing-report: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --accounts-file=$(ALL_ACCOUNTS) billing-report

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm -it housekeeper setup

test: build
	docker run --rm --entrypoint go housekeeper test -cover ./...

build-and-run: build run