IMAGE_PREFIX=koor-tech/genesis
GOLANGCI_LINT_VERSION=

TAG ?= $(shell git log -1 --pretty=%h)

build-base:
	docker build -t koor-tech/genesis-base-builder:latest -f base.Dockerfile .

build-api: build-base
	docker build  -t $(IMAGE_PREFIX)-api -f Dockerfile .
	docker tag koor-tech/genesis-api:latest koorinc/genesis:$(TAG)

build-worker: build-base
	docker build -t $(IMAGE_PREFIX)-worker -f ./cmd/worker/Dockerfile .

build-migrate: build-base
	docker build -t $(IMAGE_PREFIX)-migrate -f ./cmd/migrations/Dockerfile .

download:
	@go mod download

install-tools: download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

lint:
	@which golangci-lint > /dev/null || make install-tools
	golangci-lint run

test:
	go test ./...

push:
	docker push koorinc/genesis:$(TAG)