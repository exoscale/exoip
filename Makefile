VERSION=0.3.5
PKG=github.com/exoscale/exoip

GIMME_OS?=linux
GIMME_ARCH?=amd64


MAIN=exoip
CLI=cmd/$(MAIN)/main.go
SRCS=$(wildcard *.go)

DEST=build
BIN=$(DEST)/$(MAIN)
BINS=\
		$(BIN)        \
		$(BIN)-static

GOPATH=$(CURDIR)/.gopath
DEP=$(GOPATH)/bin/dep

export GOPATH
export PATH := $(PATH):$(GOPATH)/bin

RM?=rm -f

all: $(BIN)

$(GOPATH)/src/$(PKG):
	mkdir -p $(GOPATH)
	go get -u github.com/golang/dep/cmd/dep
	mkdir -p $(shell dirname $(GOPATH)/src/$(PKG))
	ln -sf ../../../.. $(GOPATH)/src/$(PKG)

.phony: deps
deps: $(GOPATH)/src/$(PKG)
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) ensure)

.phony: deps-status
deps-status: $(GOPATH)/src/$(PKG)
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) status)

.phony: deps-update
deps-update: deps
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) ensure -update)

.phony: generate
generate: deps
	go get -u golang.org/x/tools/cmd/stringer
	(cd $(GOPATH)/src/$(PKG) && \
		go generate)

$(BIN): $(CLI) $(SRCS)
	(cd $(GOPATH)/src/$(PKG) && \
		go build -o $@ $<)

$(BIN)-static: $(CLI) $(SRCS)
	(cd $(GOPATH)/src/$(PKG) && \
		CGO_ENABLED=0 GOOS=$(GIMME_OS) GOARCH=$(GIMME_ARCH) \
		go build -ldflags "-s" -o $@ $<)

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
	docker build --tag exoscale/exoip:$(VERSION) .

.PHONY: internal-docker
internal-docker:
	docker build --tag registry.internal.exoscale.ch/exoscale/exoip:$(VERSION) .
