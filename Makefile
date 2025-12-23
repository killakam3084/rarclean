.PHONY: help build run test clean docker-build docker-run docker-stop lint fmt

help:
	@echo "rarclean - QoL Service for TrueNAS Scale"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  fmt            - Format code"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run in Docker"
	@echo "  docker-stop    - Stop Docker container"

build:
	go build -o bin/rarclean ./cmd/rarclean

run: build
	./bin/rarclean -config config.json

test:
	go test -v -cover ./...

clean:
	rm -rf bin/ dist/
	go clean

lint:
	golangci-lint run

fmt:
	go fmt ./...

docker-build:
	docker build -t rarclean:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f rarclean
