.DEFAULT_GOAL := test
.PHONY: tidy fmt vet test test-integration test-integration-colima check ci build run clean

# Configuration
CMD_PATH := ./cmd/kafka-go-cli
BINARY_NAME := kafka-go-cli
ARGS ?=

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

test-integration-colima:
	DOCKER_HOST="unix://$(HOME)/.colima/default/docker.sock" TESTCONTAINERS_RYUK_DISABLED=true go test -v -tags=integration ./...

check: tidy vet test

ci: clean tidy fmt vet test
	@echo "✓ All checks passed! Ready for PR"

build:
	go build -o $(BINARY_NAME) $(CMD_PATH)

run: build
	./$(BINARY_NAME) $(ARGS)

clean:
	go clean
	rm -f $(BINARY_NAME)