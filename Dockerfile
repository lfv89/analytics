FROM golang:1.13-alpine

ADD . /go/src/analytics
WORKDIR /go/src/analytics

RUN go mod download
RUN go install

ENTRYPOINT /go/bin/analytics

EXPOSE 4001
