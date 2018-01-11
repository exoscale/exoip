VERSION=0.3.4
PKG=exoip

MAIN=cmd/$(PKG).go
SRCS=$(wildcard *.go)

DEST=build
BIN=$(DEST)/$(PKG)
BINS=\
		$(BIN)        \
		$(BIN)-static

RM?=rm -f
LN=ln -s

export GOPATH := $(PWD)

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

.PHONY: cleandeps
cleandeps: clean
	$(RM) -r src

.PHONY: deps
deps:
	go get github.com/exoscale/egoscale
	$(RM) src/github.com/exoscale/$(PKG)
	$(LN) $(GOPATH) src/github.com/exoscale/$(PKG)
