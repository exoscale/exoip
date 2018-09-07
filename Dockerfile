FROM golang:1.11-alpine as build

RUN mkdir -p /exoip
ADD . /exoip
WORKDIR /exoip

RUN apk add --no-cache \
            --update \
            --virtual build-dependencies \
        make \
        git \
 && cd /exoip \
 && CGO_ENABLED=0 GOOS=linux go build -mod vendor -o exoip -ldflags "-s -w" cmd/exoip


FROM linuxkit/ca-certificates:v0.6

LABEL org.label-schema.name="ExoIP" \
      org.label-schema.vendor="Exoscale" \
      org.label-schema.description="IP watchdog" \
      org.label-schema.url="https://github.com/exoscale/exoip" \
      org.label-schema.schema-version="1.0"



COPY --from=build /exoip/exoip exoip

ENTRYPOINT ["./exoip", "-O"]
