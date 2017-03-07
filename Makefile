VERSION=0.3.3-snapshot
PREFIX?=/usr/local
GOPATH=$(PWD)/build:$(PWD)
PROGRAM=exoip
GO=env GOPATH=$(GOPATH) go
SRCS=   src/exoip/network.go				\
	src/exoip/engine.go				\
	src/exoip/peer.go				\
	src/exoip/time.go				\
	src/exoip/state.go				\
	src/exoip/api.go				\
	src/exoip/metadata.go				\
	src/exoip/assert.go				\
	src/exoip/logging.go				\
	src/exoip/types.go

RM?=rm -f
LN=ln -s
MAIN=exoip.go

all: $(PROGRAM) static

$(PROGRAM): $(MAIN) $(SRCS)
	$(GO) build -o $(PROGRAM) $(MAIN)

static: $(PROGRAM)
	env CGO_ENABLED=0 GOOS=linux $(GO) build -ldflags "-s" -o $(PROGRAM)-static $(MAIN)

clean:
	$(RM) $(PROGRAM)
	$(RM) $(PROGRAM)-static
	$(RM) $(PROGRAM).asc
	$(RM) $(PROGRAM)-static.asc
	$(GO) clean

signature: all
	gpg -a --sign -u ops@exoscale.ch --detach exoip
	gpg -a --sign -u ops@exoscale.ch --detach exoip-static

cleandeps: clean
	$(RM) -r $(PWD)/build

deps:
	$(GO) get github.com/pyr/egoscale/src/egoscale
