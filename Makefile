VERSION=0.3.12-SNAPSHOT
PKG=github.com/exoscale/exoip

GIMME_OS?=linux
GIMME_ARCH?=amd64


MAIN=exoip
CLI=cmd/$(MAIN)/main.go
SRCS=$(wildcard *.go)

GOPATH=$(CURDIR)/.gopath
DEP=$(GOPATH)/bin/dep

export GOPATH
export PATH := $(PATH):$(GOPATH)/bin

RM?=rm -f

all: $(MAIN)

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

$(MAIN): $(CLI) $(SRCS)
	(cd $(GOPATH)/src/$(PKG) && \
		go build -o $@ $<)

.phony: generate
generate: deps
	go get -u golang.org/x/tools/cmd/stringer
	(cd $(GOPATH)/src/$(PKG) && \
		go generate)

clean:
	go clean

.PHONY: docker
docker:
	docker build --tag exoscale/exoip:$(VERSION) .

.PHONY: internal-docker
internal-docker:
	docker build --tag registry.internal.exoscale.ch/exoscale/exoip:$(VERSION) .
