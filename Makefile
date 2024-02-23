IMAGE_PREFIX=koor-tech/genesis

build-base:
	docker build -t koor-tech/genesis-base-builder:latest -f base.Dockerfile .

build-api: build-base
	docker build  -t $(IMAGE_PREFIX)-api -f Dockerfile .

build-worker: build-base
	docker build -t $(IMAGE_PREFIX)-worker -f ./cmd/worker/Dockerfile .

build-migrate: build-base
	docker build -t $(IMAGE_PREFIX)-migrate -f ./cmd/migrations/Dockerfile .