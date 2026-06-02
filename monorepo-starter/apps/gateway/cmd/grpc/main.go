package main

import (
	"log"

	"go-infra-monorepo-starter/apps/gateway/bootstrap"
)

func main() {
	app, err := bootstrap.NewGRPC()
	if err != nil {
		log.Fatalf("bootstrap grpc init failed: %v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("gateway grpc run failed: %v", err)
	}
}
