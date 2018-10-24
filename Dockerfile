# STEP 1 build executable binary
FROM golang:1.10-alpine3.7 as builder

ADD . $GOPATH/src/github.com/cloudtools/cloudsweeper
WORKDIR $GOPATH/src/github.com/cloudtools/cloudsweeper

RUN apk -U upgrade && \
    apk add --no-cache -U git && \
    go get ./... && \
    go test -cover ./... && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /cs cmd/cloudsweeper/*.go



FROM scratch
COPY --from=builder /cs /cs
ENTRYPOINT [ "/cs" ]
