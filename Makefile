VERSION=0.3.4-snapshot
PKG=exoip

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
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -o $@ $<

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
