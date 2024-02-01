IMAGE_NAME=koor-tech/genesis
CONTAINER_NAME=koor-genesis

all: build run

build: Dockerfile
	@echo "Building Koor platform ..."
	docker build -t $(IMAGE_NAME) .

run:
	@echo "Running Koor platform..."
	docker run -e HCLOUD_TOKEN= -d -p 8101:8000 --name $(CONTAINER_NAME) $(IMAGE_NAME)

clean:
	@echo "Cleaning up..."
	docker stop $(CONTAINER_NAME)
	docker rm $(CONTAINER_NAME)

.PHONY: all build run clean