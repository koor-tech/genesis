IMAGE_PREFIX=koor-tech/genesis

build-base:
	docker build -t koor-tech/genesis-base-builder:latest -f base.Dockerfile .

build-api:
	@make build-base
	docker build -t $(IMAGE_PREFIX)-api -f Dockerfile .

build-worker:
	@make build-base
	docker build -t $(IMAGE_PREFIX)-worker -f ./cmd/worker/Dockerfile .

build-migrate:
	docker build -t $(IMAGE_PREFIX)-migrate -f ./cmd/migrations/Dockerfile .

run:
	@echo "Running Koor platform..."
	docker run -e HCLOUD_TOKEN= -d -p 8101:8000 --name $(CONTAINER_NAME) $(IMAGE_NAME)
