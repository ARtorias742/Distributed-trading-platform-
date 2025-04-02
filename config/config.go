package config

import (
	"os"
	"strings"
)

type Config struct {
	PeerID     string
	ListenAddr string
	SeedNodes  []string
}

func LoadConfig() (*Config, error) {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		peerID = "node1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	listenAddr := ":" + port

	seedNodes := []string{}

	if seeds := os.Getenv("SEED_NODES"); seeds != "" {
		seedNodes = strings.Split(seeds, ",")
	}

	return &Config{
		PeerID:     peerID,
		ListenAddr: listenAddr,
		SeedNodes:  seedNodes,
	}, nil
}
