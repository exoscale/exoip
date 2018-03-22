FROM golang:1.10-alpine3.7 as build

LABEL org.label-schema.name="ExoIP" \
      org.label-schema.description="IP watchdog" \
      org.label-schema.url="https://github.com/exoscale/exoip" \
      org.label-schema.schema-version="1.0"

RUN mkdir -p /go/src/github.com/exoscale/exoip
ADD . /go/src/github.com/exoscale/exoip
WORKDIR /go/src/github.com/exoscale/exoip

RUN apk add --no-cache \
            --update \
            --virtual build-dependencies \
        make \
        git \
 && go get github.com/golang/dep/cmd/dep \
 && cd /go/src/github.com/exoscale/exoip \
 && dep ensure \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -o /go/bin/exoip cmd/exoip.go


FROM alpine:3.7
COPY --from=build /go/bin/exoip exoip
RUN apk add --no-cache \
        ca-certificates

ENTRYPOINT ["./exoip", "-O"]
