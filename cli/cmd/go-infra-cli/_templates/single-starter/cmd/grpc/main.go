package main

import (
	"log"

	"single-starter/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewGRPC()
	if err != nil {
		log.Fatalf("grpc bootstrap init failed: %v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("grpc server run failed: %v", err)
	}
}
