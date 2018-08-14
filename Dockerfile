FROM golang:1.10-alpine3.7

RUN apk -U upgrade && \
    apk add --no-cache -U git

RUN mkdir -p $GOPATH/src/brkt/cloudsweeper
ADD . $GOPATH/src/brkt/cloudsweeper/
WORKDIR $GOPATH/src/brkt/cloudsweeper

RUN go get ./...

RUN go build -o cloudsweeper cmd/*.go

# you need to specify the location of a file that follows the format for users as described in organizations.go.  Specify the location of a source file (maybe in an S3 bucket?) and the local path
ADD sourcepath.json ./organization.json

ENTRYPOINT [ "./cloudsweeper" ]
