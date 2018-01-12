build:
	docker build -t housekeeper .

run:
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm housekeeper

build-and-run: build run