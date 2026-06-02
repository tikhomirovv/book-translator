.PHONY: build test lint clean

BINARY := bin/translator

build:
	go build -o $(BINARY) ./cmd/translator

test:
	go test ./...

lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; skip"

clean:
	rm -rf bin/
