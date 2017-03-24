FROM alpine:3.3

RUN mkdir -p /tmp/src
ADD . /tmp/src

RUN apk --no-cache --update add --virtual build-dependencies \
      go make git \
    && cd /tmp/src \
    && make deps \
    && make \
    && cp exoip /usr/local/bin/exoip \
    && rm -r /tmp/src \
    && apk del build-dependencies

ENTRYPOINT ["/usr/local/bin/exoip", "-O"]
