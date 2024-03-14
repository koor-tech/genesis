IMAGE_PREFIX=koor-tech/genesis
GOLANGCI_LINT_VERSION=

TAG ?= $(shell git log -1 --pretty=%h)

build-base:
	docker build -t koor-tech/genesis-base-builder:latest -f base.Dockerfile .

build: build-base
	docker build  -t $(IMAGE_PREFIX) -f Dockerfile .
	docker tag koor-tech/genesis:latest koorinc/genesis:$(TAG)

download:
	@go mod download

install-tools: download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

lint:
	@which golangci-lint > /dev/null || make install-tools
	golangci-lint run ./... --timeout=5m -v

test:
	go test ./...

push:
	docker push koorinc/genesis:$(TAG)
