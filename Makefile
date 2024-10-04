.PHONY: build build-all api start task

api:
	go run cmd/api/main.go

start:
	go run cmd/indexer/main.go

task:
	go run cmd/task/usdcweth/main.go


build:
	@if [ -z "$(target)" ]; then \
		echo "target is empty, please provide target by using 'make build target={service_name}'"; \
	else \
		docker build --no-cache -t $(target) -f ./docker/$(target).Dockerfile . ; \
	fi

build-all:
	for dockerfile in ./docker/*.Dockerfile; do \
		service_name=$$(basename $$dockerfile .Dockerfile); \
		docker build --no-cache -t $$service_name -f $$dockerfile .; \
	done

docker-compose:
	docker-compose up --build