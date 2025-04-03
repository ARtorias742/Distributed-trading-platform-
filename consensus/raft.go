package consensus

import (
	"math/rand"
	"sync"
	"time"

	"github.com/artorias742/DTP/monitoring"
)

type Raft struct {
	mutex     sync.Mutex
	state     string // "follower", "candidate", " leader"
	term      int
	peers     []string
	lastHeart time.Time
	peerID    string
	votes     int
}

func NewRaft(peerID string, peers []string) *Raft {
	return &Raft{
		state:     "follower",
		term:      0,
		peers:     peers,
		lastHeart: time.Now(),
		peerID:    peerID,
		votes:     0,
	}
}

func (r *Raft) Start() {
	go r.run()
}

func (r *Raft) run() {
	logger := monitoring.GetLogger()
	rand.Seed(time.Now().UnixNano())

	for {
		r.mutex.Lock()
		currentState := r.state
		r.mutex.Unlock()

		switch currentState {
		case "follower":
			r.handleFollower()

		case "candidate":
			r.handleCandidate()

		case "leader":
			r.handleLeader()

		default:
			logger.Error("Unknown state", "state", currentState)
			time.Sleep(1 * time.Second)
		}
	}
}

// // handleFollower manages the follower state, waiting for heartbeats or triggering an election.
func (r *Raft) handleFollower() {
	logger := monitoring.GetLogger()

	// Random electionn timeout between 150-300ms
	timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
	time.Sleep(timeout)

	r.mutex.Lock()
	if time.Since(r.lastHeart) > timeout {
		logger.Info("No heartbeat received, becoming candidate", "peerID", r.peerID)
		r.state = "candidate"
		r.term++
		r.votes = 1
		r.lastHeart = time.Now()
	}
	r.mutex.Unlock()
}

func (r *Raft) handleCandidate() {
	logger := monitoring.GetLogger()
	logger.Info("Starting election", "peerID", r.peerID, "term", r.term)

	majority := (len(r.peers)+1)/2 + 1 // +1 for self
	voteChan := make(chan bool, len(r.peers))

	// Simulate sending vote requests to peers
	for _, peer := range r.peers {
		go func(peer string) {
			// In a real system, send RPC to peer and wait for response
			// Here, simulate with 50% chance of vote
			vote := rand.Intn(2) == 1
			voteChan <- vote
		}(peer)
	}

	// Collect votes
	timeout := time.After(200 * time.Millisecond)
	for i := 0; i < len(r.peers); i++ {
		select {
		case vote := <-voteChan:
			r.mutex.Lock()
			if vote {
				r.votes++
			}
			if r.votes >= majority && r.state == "candidate" {
				logger.Info("Won election, becoming leader", "peerID", r.peerID, "term", r.term)
				r.state = "leader"
			}
			r.mutex.Unlock()
		case <-timeout:
			logger.Info("Election timeout, reverting to follower", "peerID", r.peerID)
			r.mutex.Lock()
			r.state = "follower"
			r.votes = 0
			r.mutex.Unlock()
			return
		}
	}

	// If no majority, revert to follower
	r.mutex.Lock()
	if r.state == "candidate" {
		logger.Info("No majority, reverting to follower", "peerID", r.peerID)
		r.state = "follower"
		r.votes = 0
	}
	r.mutex.Unlock()
}

func (r *Raft) handleLeader() {
	logger := monitoring.GetLogger()
	logger.Info("Acting as leader", "peerID", r.peerID, "term", r.term)

	// Send heartbeats every 50ms
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		if r.state != "leader" {
			r.mutex.Unlock()
			return
		}
		r.mutex.Unlock()

		// Simulate sending heartbeats to peers
		for _, peer := range r.peers {
			go func(peer string) {
				// In a real system, send heartbeat RPC
				// Here, just log it
				logger.Debug("Sent heartbeat", "to", peer, "term", r.term)
				r.mutex.Lock()
				r.lastHeart = time.Now() // Update own heartbeat time
				r.mutex.Unlock()
			}(peer)
		}

		// Check if we should step down (simplified)
		// In a real system, this would be based on receiving higher term from peers
		if rand.Intn(100) < 5 { // 5% chance to simulate term conflict
			r.mutex.Lock()
			logger.Info("Stepping down due to simulated term conflict", "peerID", r.peerID)
			r.state = "follower"
			r.votes = 0
			r.term++
			r.mutex.Unlock()
			return
		}
	}
}
