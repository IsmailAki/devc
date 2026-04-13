.PHONY: build install clean test

BINARY_NAME=devc
GO=go

build:
	$(GO) build -o bin/$(BINARY_NAME) .

install: build
	cp bin/$(BINARY_NAME) /usr/local/bin/

clean:
	rm -rf bin/

test:
	$(GO) test -v ./...

run:
	$(GO) run . $(ARGS)

lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

all: tidy fmt build
