.PHONY: all build test clean cover

all: build test

build:
	go build -v ./...

test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

clean:
	go clean
	rm -f coverage.out coverage.html
