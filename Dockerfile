FROM golang:1.10-alpine3.7

RUN apk -U upgrade && \
    apk add --no-cache -U git

RUN mkdir -p $GOPATH/src/brkt/olga
ADD . $GOPATH/src/brkt/olga/
WORKDIR $GOPATH/src/brkt/olga

RUN go get ./...

RUN go build -o olga cmd/*.go

ADD https://s3-us-west-2.amazonaws.com/packages.int.brkt.net/org/latest/organization.json ./organization.json

ENTRYPOINT [ "./olga" ]