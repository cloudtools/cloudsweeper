FROM golang:1.10-alpine3.7

RUN apk -U upgrade && \
    apk add --no-cache -U git

RUN mkdir -p $GOPATH/src/github.com/cloudtools/cloudsweeper
ADD . $GOPATH/src/github.com/cloudtools/cloudsweeper/
WORKDIR $GOPATH/src/github.com/cloudtools/cloudsweeper

RUN go get ./...

RUN go build -o cs cmd/cloudsweeper/*.go

# you need to specify the location of a file that follows the format for users
# as described in organizations.go.  Specify the location of a source file (can
# be either a local path or a URL to e.g. an S3 bucket).
ADD example-org.json ./organization.json

ENTRYPOINT [ "./cs" ]
