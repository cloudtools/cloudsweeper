build:
	docker build -t housekeeper .

run:
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper


cleanup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper --cleanup

review: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e SMTP_USER \
		-e SMTP_PASS \
		--rm housekeeper --review

setup: build
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm -it housekeeper --setup

build-and-run: build run