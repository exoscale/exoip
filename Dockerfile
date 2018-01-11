FROM golang:1.9-alpine3.7 as build

RUN mkdir -p /app
ADD . /app
WORKDIR /app

RUN apk add --no-cache \
            --update \
            --virtual build-dependencies \
        make \
        git \
 && cd /app \
 && make deps \
 && make build/exoip-static


FROM alpine:3.7
COPY --from=build /app/build/exoip-static exoip
RUN apk add --no-cache \
        ca-certificates \
        iproute2

ENTRYPOINT ["./exoip", "-O"]
