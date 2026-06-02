package main

import (
	"log"

	"go-infra-monorepo-starter/apps/gateway/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatalf("bootstrap init failed: %v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("gateway run failed: %v", err)
	}
}
