GO = go
GOTESTSUM = gotestsum
GOFMT = gofmt
GOLANGCILINT=golangci-lint -vv
GOSEC=gosec

export GO111MODULE = on
GO_FLAGS =

DOCKER = docker
REGISTRY ?= docker.io
INSECURE_REGISTRY = false

GO_PACKAGES = ./...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

COMMIT = $(shell git rev-parse HEAD)
TAG = $(shell git describe --exact-match --abbrev=0 --tags '$(COMMIT)' 2> /dev/null || true)
DIRTY = $(shell git diff --shortstat 2> /dev/null | tail -n1)

# Use a tag if set, otherwise use the commit hash
ifeq ($(TAG),)
VERSION := $(COMMIT)
else
VERSION := $(TAG)
endif

CONTROLLER_IMAGE = $(REGISTRY)/itsthatdood/namespaced-route-validator
ifneq ($(TAG),)
CONTROLLER_IMAGE_TAGGED := $(CONTROLLER_IMAGE):$(VERSION)
endif
CONTROLLER_IMAGE_LATEST = $(CONTROLLER_IMAGE):latest

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)

# Check for changed files
ifneq ($(DIRTY),)
VERSION := $(VERSION)+dirty
endif

GO_LD_FLAGS = -X main.VERSION=$(VERSION)

all: controller

controller: $(GO_FILES)
	GOOS=$(2) CGO_ENABLED=0 $(GO) build -o $@ -installsuffix cgo $(GO_FLAGS) -ldflags "$(GO_LD_FLAGS)" ./cmd/controller

test:
	$(GOTESTSUM) $(GO_FLAGS) --junitfile report.xml --format testname -- "-coverprofile=coverage.out" $(GO_PACKAGES)

fmt:
	$(GOFMT) -s -w $(GO_FILES)

lint:
	$(GOLANGCILINT) run --enable goimports --timeout=5m

lint-gosec:
	$(GOSEC) -r -severity low -exclude-generated

clean:
	$(RM) ./controller
	$(RM) *-static*
	$(RM) controller*.yaml
	$(RM) controller.image*
	$(RM) -r ./dist

docker-build: docker/controller.Dockerfile controller
	mkdir -p dist
	cp controller dist/
	$(DOCKER) build -t $(CONTROLLER_IMAGE_LATEST) -f docker/controller.Dockerfile .

docker-push: clean docker-build
	if [ -n "$(TAG)" ]; then \
		docker tag $(CONTROLLER_IMAGE_LATEST) $(CONTROLLER_IMAGE_TAGGED); \
		docker push $(CONTROLLER_IMAGE_TAGGED); \
	fi
	docker push $(CONTROLLER_IMAGE_LATEST)