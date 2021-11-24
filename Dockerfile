FROM golang:1.13.4-alpine3.10 AS build

LABEL maintainer="nenad.zaric@trickest.com"

RUN apk add --no-cache --upgrade git openssh-client ca-certificates bash

COPY . /app

WORKDIR /app

RUN go build -o mkpath

FROM alpine:3.10

RUN apk add bash

RUN mkdir -p /hive/in

RUN mkdir -p /hive/out

COPY --from=build /app/mkpath /usr/bin/mkpath

ENTRYPOINT ["mkpath"]