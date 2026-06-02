.PHONY: tidy test run

tidy:
	go work sync
	cd apps/gateway && go mod tidy

test:
	cd apps/gateway && go test ./...

run:
	cd apps/gateway && go run ./cmd

