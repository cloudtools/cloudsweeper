FROM alpine:3.7 AS organization
RUN apk -U upgrade && \
    apk add --no-cache -U git
ARG CACHE_DATE=a_date
RUN git clone https://jenkins-ro:YrerrGLoNE9fcZ9Vn99YHqrN@gerrit.int.brkt.net/a/org /src/org && \
    cp /src/org/organization.json /src/organization.json && \
    rm -rf /src/org

FROM golang:1.9-alpine3.7

RUN apk -U upgrade && \
    apk add --no-cache -U git

RUN mkdir -p $GOPATH/src/brkt/olga
ADD . $GOPATH/src/brkt/olga/
WORKDIR $GOPATH/src/brkt/olga

RUN go get ./...

COPY --from=organization /src/organization.json ./organization.json
RUN go build -o olga cmd/*.go
ENTRYPOINT [ "./olga" ]