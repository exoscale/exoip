FROM golang:1.9-alpine as build

LABEL org.label-schema.name="ExoIP" \
      org.label-schema.description="IP watchdog" \
      org.label-schema.url="https://github.com/exoscale/exoip" \
      org.label-schema.schema-version="1.0"

RUN mkdir -p /go/src/github.com/exoscale/exoip
ADD . /go/src/github.com/exoscale/exoip
WORKDIR /go/src/github.com/exoscale/exoip

RUN apk --no-cache \
        --update add \
        --virtual build-dependencies \
        make git \
 && cd /go/src/github.com/exoscale/exoip \
 && go get -u github.com/golang/dep/cmd/dep \
 && dep ensure \
 && make build/exoip-static


FROM linuxkit/ca-certificates:de21b84d9b055ad9dcecc57965b654a7a24ef8e0-amd64
COPY --from=build /go/src/github.com/exoscale/exoip/build/exoip-static exoip
ENTRYPOINT ["./exoip", "-O"]
