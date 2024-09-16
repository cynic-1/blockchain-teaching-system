package main

import (
	"github.com/cynic-1/blockchain-teaching-system/internal/config"
	"github.com/cynic-1/blockchain-teaching-system/internal/server"
	"log"
)

func main() {
	cfg := config.NewConfig()
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Starting server on %s", cfg.ServerPort)
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
