.PHONY: build test lint clean install

BIN_DIR := bin
BIN := $(BIN_DIR)/codely

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/codely

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f $(BIN)

install: build
	mkdir -p ~/.local/bin
	cp $(BIN) ~/.local/bin/codely
