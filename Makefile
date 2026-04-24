.PHONY: build test lint clean install

BUILD_DIR := bin
BINARY    := $(BUILD_DIR)/figaro

build:
	go build -o $(BINARY) ./cmd/figaro/

test:
	go test ./... -v -count=1

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

install:
	go install ./cmd/figaro/
