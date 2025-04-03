package main

import (
	"net/http"
	"time"

	"github.com/artorias742/DTP/api"
	"github.com/artorias742/DTP/config"
	"github.com/artorias742/DTP/monitoring"
	"github.com/artorias742/DTP/network"
	"github.com/artorias742/DTP/recovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize structured logging
	monitoring.InitLogging()
	logger := monitoring.GetLogger()
	logger.Info("Starting distributed trading platform", "date", time.Now().Format(time.RFC3339))

	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}
	logger.Info("Configuration loaded",
		"peerID", cfg.PeerID,
		"listenAddr", cfg.ListenAddr,
		"seedNodes", cfg.SeedNodes)

	// Initialize Prometheus metrics
	monitoring.InitMetrics()
	logger.Info("Metrics initialized")

	// Start metrics server on :9090
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("Starting metrics server", "addr", ":9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Error("Metrics server failed", "error", err)
		}
	}()

	// Initialize peer node
	peer := network.NewPeer(cfg)

	// Start API server for user interaction
	apiServer := api.NewServer(peer)
	go apiServer.Start()

	// Start peer with recovery mechanism
	if err := recovery.StartWithRecovery(peer); err != nil {
		logger.Fatal("Failed to start peer with recovery", "error", err)
	}

	// Keep the main goroutine running
	select {}

}
