VERSION=0.3.5-snapshot
PKG=exoip

GIMME_OS?=linux
GIMME_ARCH?=amd64

MAIN=cmd/$(PKG).go
SRCS=$(wildcard *.go)

DEST=build
BIN=$(DEST)/$(PKG)
BINS=\
		$(BIN)        \
		$(BIN)-static

RM?=rm -f

all: $(BIN)

$(BIN): $(MAIN) $(SRCS)
	go build -o $@ $<

$(BIN)-static: $(MAIN) $(SRCS)
	CGO_ENABLED=0 GOOS=$(GIMME_OS) GOARCH=$(GIMME_ARCH) go build -ldflags "-s" -o $@ $<

clean:
	$(RM) -r $(DEST)
	go clean

.PHONY: signature
signature: $(BINS)
	$(foreach bin,$^,\
		$(RM) $(bin).asc; \
		gpg -a --sign -u ops@exoscale.ch --detach $(bin);)

.PHONY: docker
docker:
	docker build --tag exoscale/exoip:$(shell git rev-parse --short HEAD) .
