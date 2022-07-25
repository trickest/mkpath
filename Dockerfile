FROM golang:1.17.2-alpine AS build-env

RUN apk add git
ADD . /go/src/mkpath
WORKDIR /go/src/mkpath
RUN git checkout 1fe6937 && go build -o mkpath

FROM alpine:3.14
LABEL licenses.mkpath.name="MIT" \
      licenses.mkpath.url="https://github.com/trickest/mkpath/blob/1fe6937da4346340b514759e83a40ba231bba5e2/LICENSE" \
      licenses.golang.name="bsd-3-clause" \
      licenses.golang.url="https://go.dev/LICENSE?m=text"

COPY --from=build-env /go/src/mkpath/mkpath /bin/mkpath

RUN mkdir -p /hive/in /hive/out

WORKDIR /app
RUN apk add bash

ENTRYPOINT [ "mkpath" ]
