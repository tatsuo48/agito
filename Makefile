.PHONY: build test lint clean install

BUILD_DIR := bin
BINARY    := $(BUILD_DIR)/agito

build:
	go build -o $(BINARY) ./cmd/agito/

test:
	go test ./... -v -count=1

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

install:
	go install ./cmd/agito/
