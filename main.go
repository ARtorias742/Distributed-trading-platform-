package main

import (
	"log"

	"github.com/ARtorias742/DTP/config"
	"github.com/ARtorias742/DTP/network"
)

func main() {

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize and start peer node
	peer := network.NewPeer(cfg)
	if err := peer.Start(); err != nil {
		log.Fatalf("Failed to start peer: %v", err)
	}

	// Keep the application running
	select {}

}
