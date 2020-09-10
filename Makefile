# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build

VERSION=0.9.2
GOLANG_VERSION?=1.14
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

REPO_DIR:=$(shell pwd)
ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/wavefront.XXXXXX)
endif

DOCKER_REPO=wavefronthq
DOCKER_IMAGE=wavefront-fargate-collector

# for dev testing, the built image will also be tagged with this name
OVERRIDE_IMAGE_NAME?=$(FARGATE_COLLECTOR_TEST_IMAGE)

LDFLAGS=-w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT)

# Binary names
BINARY_NAME=fargate_collector
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_DARWIN=$(BINARY_NAME)_darwin

all:  build-linux

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_LINUX) -v
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_DARWIN) -v

container:
	# Run build in a container in order to have reproducible builds
	docker run --rm -v $(TEMP_DIR):/build -v $(REPO_DIR):/go/src/github.com/wavefronthq/wavefront-fargate-collector -w /go/src/github.com/wavefronthq/wavefront-fargate-collector golang:$(GOLANG_VERSION) /bin/bash -c "\
		GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags \"$(LDFLAGS)\" -a -tags netgo -o /build/$(BINARY_NAME)_linux github.com/wavefronthq/wavefront-fargate-collector/"

	cp Dockerfile $(TEMP_DIR)
	cp -r static $(TEMP_DIR)
	docker build --pull -t $(DOCKER_REPO)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	docker tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif
