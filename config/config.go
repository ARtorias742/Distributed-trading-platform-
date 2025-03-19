package config

type Config struct {
	PeerID     string
	ListenAddr string
	SeedNodes  []string
}

func LoadConfig() (*Config, error) {
	// In a real implementation, this would load from file or env vars

	return &Config{
		PeerID:     "node1",
		ListenAddr: ":8080",
		SeedNodes:  []string{":8081", ":8082"},
	}, nil
}
