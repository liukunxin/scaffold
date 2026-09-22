.PHONY: build run run-grpc test vet fmt tidy clean

build:
	go build ./...

run:
	go run ./cmd/http

run-grpc:
	go run ./cmd/grpc

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	go clean -testcache
