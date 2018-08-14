FROM golang:1.10-alpine3.7

RUN apk -U upgrade && \
    apk add --no-cache -U git

RUN mkdir -p $GOPATH/src/brkt/olga
ADD . $GOPATH/src/brkt/olga/
WORKDIR $GOPATH/src/brkt/olga

RUN go get ./...

RUN go build -o olga cmd/*.go

# you need to specify the location of a file that follows the format for users as described in organizations.go.  Specify the location of a source file (maybe in an S3 bucket?) and the local path
ADD sourcepath.json ./organization.json

ENTRYPOINT [ "./olga" ]
