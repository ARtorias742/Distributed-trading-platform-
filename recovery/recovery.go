package recovery

import (
	"time"

	"github.com/artorias742/DTP/monitoring"
	"github.com/artorias742/DTP/network"
)

func StartWithRecovery(peer *network.Peer) error {
	logger := monitoring.GetLogger()
	for {
		err := peer.Start()
		if err != nil {
			logger.Error("Peer crashed", "error", err)
			time.Sleep(5 * time.Second) // Simple backoff
			continue
		}
		return nil
	}
}
