# parent image
FROM golang:1.17.0-alpine3.14

WORKDIR /go/src/app

ADD . /go/src/app

RUN go get -u

# build executable
RUN go build
