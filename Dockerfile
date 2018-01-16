FROM golang:1.9-alpine3.7

# Install packages for python
RUN apk -U upgrade && \
    apk add --no-cache -U python-dev py-pip git python

# Install dependencies for python
COPY requirements.txt /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt

RUN mkdir -p $GOPATH/src/brkt/housekeeper
ADD . $GOPATH/src/brkt/housekeeper/
WORKDIR $GOPATH/src/brkt/housekeeper

RUN go get ./...

RUN python accounts_retriever.py --output=$GOPATH/src/brkt/housekeeper/aws_accounts.json
RUN go build -o go-housekeeper cmd/*.go
ENTRYPOINT [ "./go-housekeeper" ]