ACCOUNTS_FROM_MV 	:= aws_accounts.json
FRIENDLIES 			:= friendly_accounts.json
ACCOUNTS_FILE 		:= $(FRIENDLIES)
WARNING_HOURS		:= 48

build:
	docker build -t housekeeper .

run:
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --accounts-file=$(ACCOUNTS_FILE)

cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --cleanup --accounts-file=$(ACCOUNTS_FILE)

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --review --accounts-file=$(ACCOUNTS_FILE)

mark: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --mark-for-cleanup --accounts-file=$(ACCOUNTS_FILE)

warn: build 
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --warning --warning-hours=$(WARNING_HOURS) --accounts-file=$(ACCOUNTS_FILE)

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm -it housekeeper --setup

test: build
	docker run --rm --entrypoint go housekeeper test ./...

build-and-run: build run