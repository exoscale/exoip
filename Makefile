VERSION=0.3.0-snapshot
PREFIX?=/usr/local
GOPATH=$(PWD)/build:$(PWD)
PROGRAM=exoip
GO=env GOPATH=$(GOPATH) go
SRCS=   src/exoip/network.go				\
	src/exoip/engine.go				\
	src/exoip/peer.go				\
	src/exoip/time.go				\
	src/exoip/state.go				\
	src/exoip/types.go

RM?=rm -f
LN=ln -s
MAIN=exoip.go

all: $(PROGRAM)

$(PROGRAM): $(MAIN) $(SRCS)
	$(GO) build -o $(PROGRAM) $(MAIN)

clean:
	$(RM) $(PROGRAM)
	$(GO) clean

cleandeps: clean
	$(RM) $(PWD)/build

#deps:
#	$(GO) get gopkg.in/yaml.v2
