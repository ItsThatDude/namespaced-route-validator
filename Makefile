GO = go
GOTESTSUM = gotestsum
GOFMT = gofmt
GOLANGCILINT=golangci-lint -vv
GOSEC=gosec

export GO111MODULE = on
GO_FLAGS =

DOCKER = docker
REGISTRY ?= docker.io
CONTROLLER_IMAGE = $(REGISTRY)/itsthatdood/namespaced-route-validator:latest
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

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)

# Check for changed files
ifneq ($(DIRTY),)
VERSION := $(VERSION)+dirty
endif

GO_LD_FLAGS = -X main.VERSION=$(VERSION)

all: controller

controller: $(GO_FILES)
	$(GO) build -o $@ $(GO_FLAGS) -ldflags "$(GO_LD_FLAGS)" ./cmd/controller

define binary
$(1)-static-$(2)-$(3): $(GO_FILES)
	GOOS=$(2) GOARCH=$(3) CGO_ENABLED=0 $(GO) build -o $$@ -installsuffix cgo $(GO_FLAGS) -ldflags "$(GO_LD_FLAGS)" ./cmd/$(1)
endef

define binaries
$(call binary,controller,$1,$2)
endef

$(eval $(call binaries,linux,amd64))
$(eval $(call binaries,linux,arm64))
$(eval $(call binaries,linux,arm))
$(eval $(call binaries,darwin,amd64))

controller-static: controller-static-$(GOOS)-$(GOARCH)
	cp $< $@

define image
$(1).image.$(3)-$(4): docker/$(1).Dockerfile $(1)-static-$(3)-$(4)
	mkdir -p dist/$(1)_$(3)_$(4)
	cp $(1)-static-$(3)-$(4) dist/$(1)_$(3)_$(4)/$(1)
	$(DOCKER) build --build-arg TARGETARCH=$(4) -t $(2)-$(3)-$(4) -f docker/$(1).Dockerfile .
	@echo $(2)-$(3)-$(4) >$$@.tmp
	@mv $$@.tmp $$@
endef

define images
$(call image,controller,${CONTROLLER_IMAGE},$1,$2)
endef

$(eval $(call images,linux,amd64))
$(eval $(call images,linux,arm64))
$(eval $(call images,linux,arm))

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

#push-controller: clean controller.image.$(OS)-$(ARCH)
#	docker tag $(CONTROLLER_IMAGE)-$(OS)-$(ARCH) $(CONTROLLER_IMAGE)
push-controller: clean controller.image.linux-amd64
	docker tag $(CONTROLLER_IMAGE)-linux-amd64 $(CONTROLLER_IMAGE)
	docker push $(CONTROLLER_IMAGE)