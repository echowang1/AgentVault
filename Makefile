.PHONY: all build test clean docker-build lint format

all: build

build:
	@echo "Building server..."
	cd server && go build -o ../bin/server ./cmd/server
	@echo "Building SDK..."
	cd sdk && npm run build

test:
	@echo "Running server tests..."
	cd server && go test -v ./...
	@echo "Running SDK tests..."
	cd sdk && npm test

clean:
	rm -rf bin/
	cd sdk && rm -rf dist/

docker-build:
	docker build -t agent-mpc-wallet:latest -f docker/Dockerfile .

lint-server:
	cd server && golangci-lint run

lint-sdk:
	cd sdk && npm run lint

format-server:
	cd server && gofmt -s -w .

format-sdk:
	cd sdk && npm run format
