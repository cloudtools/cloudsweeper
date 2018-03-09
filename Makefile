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
		--rm housekeeper --cleanup --accounts-file=$(FRIENDLIES)

reset: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --reset --accounts-file=$(FRIENDLIES)

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --review --accounts-file=$(FRIENDLIES)

mark: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --mark-for-cleanup --accounts-file=$(FRIENDLIES)

warn: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --warning --warning-hours=$(WARNING_HOURS) --accounts-file=$(FRIENDLIES)

billing-report: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --billing-report --accounts-file=$(ALL_ACCOUNTS)

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm -it housekeeper --setup

test: build
	docker run --rm --entrypoint go housekeeper test -cover ./...

build-and-run: build run