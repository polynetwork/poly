GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --always --tags --long)
BUILD_NODE_PAR = -ldflags "-X github.com/polynetwork/poly/common/config.Version=$(VERSION)" #-race

ARCH=$(shell uname -m)
DBUILD=docker build
DRUN=docker run
DOCKER_NS ?= polynetwork
DOCKER_TAG=$(ARCH)-$(VERSION)

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)
TOOLS=./tools
ABI=$(TOOLS)/abi
NATIVE_ABI_SCRIPT=./cmd/abi/native_abi_script

poly: $(SRC_FILES)
	$(GC)  $(BUILD_NODE_PAR) -o poly main.go
 
sigsvr: $(SRC_FILES) abi 
	$(GC)  $(BUILD_NODE_PAR) -o sigsvr sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr $(TOOLS)

abi: 
	@if [ ! -d $(ABI) ];then mkdir -p $(ABI) ;fi
	@cp $(NATIVE_ABI_SCRIPT)/*.json $(ABI)

tools: sigsvr abi

all: poly tools

poly-cross: poly-windows poly-linux poly-darwin

poly-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o poly-windows-amd64.exe main.go

poly-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o poly-linux-amd64 main.go

poly-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o poly-darwin-amd64 main.go

tools-cross: tools-windows tools-linux tools-darwin

tools-windows: abi 
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-windows-amd64.exe sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-windows-amd64.exe $(TOOLS)

tools-linux: abi 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-linux-amd64 sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-linux-amd64 $(TOOLS)

tools-darwin: abi 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-darwin-amd64 sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-darwin-amd64 $(TOOLS)

all-cross: poly-cross tools-cross abi

format:
	$(GOFMT) -w main.go

#docker/payload: docker/build/bin/poly docker/DockerfileWithConfig
#	@echo "Building poly payload"
#	@mkdir -p $@
#	@cp docker/Dockerfile $@
#	@cp docker/build/bin/poly $@
#	@touch $@
#
#docker/build/bin/%: Makefile
#	@echo "Building poly in docker"
#	@mkdir -p docker/build/bin docker/build/pkg
#	@$(DRUN) --rm \
#		-v $(abspath docker/build/bin):/go/bin \
#		-v $(abspath docker/build/pkg):/go/pkg \
#		-v $(GOPATH)/src:/go/src \
#		-w /go/src/github.com/polynetwork/poly \
#		golang:1.9.5-stretch \
#		$(GC)  $(BUILD_NODE_PAR) -o docker/build/bin/poly main.go
#	@touch $@

docker/withConfig: Makefile
	@echo "Building poly docker image with configuration"
	@$(DBUILD) --no-cache -t $(DOCKER_NS)/poly:$(DOCKER_TAG) -f docker/DockerfileWithConfig ./

dockerImg: Makefile
	@echo "Building poly docker image"
	@$(DBUILD) --no-cache -t $(DOCKER_NS)/poly:$(DOCKER_TAG) - < docker/Dockerfile

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -rf poly poly-* tools docker/payload docker/build

