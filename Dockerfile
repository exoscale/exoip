FROM golang:1.10-alpine as build

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
 && dep ensure -v -vendor-only \
 && CGO_ENABLED=0 GOOS=linux go install -ldflags "-s" github.com/exoscale/exoip/cmd/exoip


FROM linuxkit/ca-certificates:v0.4
COPY --from=build /go/bin/exoip exoip

ENTRYPOINT ["./exoip", "-O"]
