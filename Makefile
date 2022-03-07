GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --always --tags --long)

TOP:=$(realpath .)/temp
export CGO_CFLAGS:=-I$(TOP)/bls/include -I$(TOP)/mcl/include -I/usr/local/opt/openssl/include
export CGO_LDFLAGS:=-L$(TOP)/bls/lib -L/usr/local/opt/openssl/lib
export LD_LIBRARY_PATH:=$(TOP)/bls/lib:$(TOP)/mcl/lib:/usr/local/opt/openssl/lib:/usr/local/opt/gmp/lib
export LIBRARY_PATH:=$(LD_LIBRARY_PATH)
export DYLD_FALLBACK_LIBRARY_PATH:=$(LD_LIBRARY_PATH)
export GO111MODULE:=on

ARCH=$(shell uname -m)
DBUILD=docker build
DRUN=docker run
DOCKER_NS ?= polynetwork
DOCKER_TAG=$(ARCH)-$(VERSION)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	BUILD_NODE_PAR = -tags netgo -ldflags '-w -extldflags "-static -lm" -X github.com/polynetwork/poly/common/config.Version=$(VERSION) -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn' #-race
else
	BUILD_NODE_PAR = -tags netgo -ldflags '-X github.com/polynetwork/poly/common/config.Version=$(VERSION) -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn' #-race
endif

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)
TOOLS=./tools
ABI=$(TOOLS)/abi
NATIVE_ABI_SCRIPT=./cmd/abi/native_abi_script

deps:
	./scripts/install_dependencies.sh

poly: $(SRC_FILES) deps
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

poly-windows: deps
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o poly-windows-amd64.exe main.go

poly-linux: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o poly-linux-amd64 main.go

poly-darwin: deps
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

