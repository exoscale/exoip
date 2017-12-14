FROM golang:1.9-alpine
RUN mkdir -p /app
ADD . /app
WORKDIR /app

RUN apk --no-cache \
        --update add \
        --virtual build-dependencies \
        make git \
 && cd /app \
 && make deps \
 && make

FROM linuxkit/ca-certificates:de21b84d9b055ad9dcecc57965b654a7a24ef8e0
COPY --from=0 /app/build/exoip .
ENTRYPOINT ["./exoip", "-O"]
