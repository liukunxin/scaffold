package main

import (
	"log"

	"monorepo-starter/services/gateway/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewGRPC()
	if err != nil {
		log.Fatalf("grpc bootstrap init failed: %v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("gateway grpc run failed: %v", err)
	}
}
