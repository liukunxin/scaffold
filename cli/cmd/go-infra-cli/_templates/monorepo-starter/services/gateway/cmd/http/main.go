package main

import (
	"log"

	"monorepo-starter/services/gateway/internal/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatalf("bootstrap init failed: %v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("gateway http run failed: %v", err)
	}
}
