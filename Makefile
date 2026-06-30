.PHONY: build run test vet fmt lint

build:
	go build ./...

run:
	go run ./cmd/jungle

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run
