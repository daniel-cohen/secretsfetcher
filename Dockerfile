FROM golang:1.15.7-alpine3.13 as builder
#FROM golang:1.15-buster as builder

LABEL maintainer=daniel.cohen@nearmap.com

ARG VERSION=unset

RUN apk update && apk add git && apk add ca-certificates
#RUN apt-get update && apt-get install ca-certificates

RUN adduser -D -g '' appuser

ENV GO111MODULE=on

WORKDIR /code
COPY . /code

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w -X 'main.Version=$VERSION'" -o /go/bin/secretsfetcher .

#FROM debian:buster
FROM alpine:3.13

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/secretsfetcher /app/secretsfetcher


#RUN mkdir -p /secretsfetcher/secrets && chown -R appuser /secretsfetcher/secrets

USER appuser
#VOLUME /secretsfetcher/secrets

WORKDIR /app

ENTRYPOINT ["/app/secretsfetcher"]