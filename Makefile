.PHONY: help build test bench lint clean fmt vet tidy run

.DEFAULT_GOAL := build

help:
	@echo "  build    - Build the tacktack binary"
	@echo "  install  - Install binary to GOBIN (~/go/bin)"
	@echo "  run      - Run the tacktack"
	@echo "  test     - Run tests with race detector"
	@echo "  bench    - Run benchmark tests"
	@echo "  cover    - Generate HTML coverage report"
	@echo "  lint     - Run golangci-lint"
	@echo "  fmt      - Format code with gofmt"
	@echo "  vet      - Run go vet"
	@echo "  tidy     - Run go mod tidy"
	@echo "  clean    - Remove build artifacts"

build:
	@mkdir -p bin
	go build -o ./bin/tacktack ./cmd/tacktack

install: tidy
	go install ./cmd/tacktack

run: build
	./bin/tacktack

test:
	go test -v -race -coverprofile=coverage.out ./...

bench:
	go test -bench=. -benchmem ./internal/tacktack

cover:
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f ./bin/tacktack coverage.out coverage.html
