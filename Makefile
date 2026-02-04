.PHONY: build test lint clean install

build:
	go build -o codely ./cmd/codely

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f codely

install: build
	mkdir -p ~/.local/bin
	cp codely ~/.local/bin/codely
