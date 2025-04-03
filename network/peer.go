package network

import (
	"bufio"
	"crypto/rand"
	"fmt"

	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/artorias742/DTP/config"
	"github.com/artorias742/DTP/consensus"
	"github.com/artorias742/DTP/monitoring"
	"github.com/artorias742/DTP/security"
	"github.com/artorias742/DTP/trading"
)

type Peer struct {
	config    *config.Config
	OrderBook *trading.OrderBook
	auth      *security.AuthManager
	raft      *consensus.Raft
	listener  net.Listener
	peers     map[string]net.Conn
}

func NewPeer(cfg *config.Config) *Peer {
	return &Peer{
		config:    cfg,
		OrderBook: trading.NewOrderBook(),
		auth:      security.NewAuthManager([]byte("32-byte-secret-key-here!!")), // Must be 32 bytes for AES-256
		raft:      consensus.NewRaft(cfg.PeerID, cfg.SeedNodes),
		peers:     make(map[string]net.Conn),
	}
}

func (p *Peer) Start() error {
	logger := monitoring.GetLogger()
	logger.Info("Starting peer", "addr", p.config.ListenAddr)

	listener, err := net.Listen("tcp", p.config.ListenAddr)
	if err != nil {
		return err
	}
	p.listener = listener

	// Start Raft consensus
	p.raft.Start()

	// Start accepting connections
	go p.acceptConnections()

	// Connect to seed nodes
	p.connectToSeeds()

	logger.Info("Peer started successfully", "addr", p.config.ListenAddr)
	return nil
}

func (p *Peer) acceptConnections() {
	logger := monitoring.GetLogger()
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection", "error", err)
			continue
		}
		go p.handleConnection(conn)
	}
}

func (p *Peer) handleConnection(conn net.Conn) {
	logger := monitoring.GetLogger()
	defer conn.Close()

	if !p.authenticate(conn) {
		logger.Warn("Authentication failed", "remote", conn.RemoteAddr())
		return
	}

	logger.Info("Connection authenticated", "remote", conn.RemoteAddr())
	for {
		msg, err := p.readMessage(conn)
		if err != nil {
			logger.Error("Message handling failed", "error", err)
			return
		}
		if err := p.processMessage(msg); err != nil {
			logger.Warn("Message processing failed", "error", err)
		}
	}
}

func (p *Peer) connectToSeeds() {
	logger := monitoring.GetLogger()
	for _, addr := range p.config.SeedNodes {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			logger.Warn("Failed to connect to seed", "addr", addr, "error", err)
			continue
		}
		p.peers[addr] = conn
		logger.Info("Connected to seed", "addr", addr)
	}
}

// authenticate performs a handshake with the connecting peer using ECDSA signatures.
// It sends a challenge, receives a signed response, and verifies it.
func (p *Peer) authenticate(conn net.Conn) bool {
	logger := monitoring.GetLogger()

	// Generate a random challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		logger.Error("Failed to generate challenge", "error", err)
		return false
	}

	// Send challenge to the peer
	if _, err := conn.Write(challenge); err != nil {
		logger.Error("Failed to send challenge", "error", err)
		return false
	}

	// Read signature (64 bytes for P-256 ECDSA) and public key (64 bytes raw X||Y)
	reader := bufio.NewReader(conn)
	signature := make([]byte, 64)
	if _, err := io.ReadFull(reader, signature); err != nil {
		logger.Error("Failed to read signature", "error", err)
		return false
	}

	pubKey := make([]byte, 64)
	if _, err := io.ReadFull(reader, pubKey); err != nil {
		logger.Error("Failed to read public key", "error", err)
		return false
	}

	// Verify the signature
	if !security.VerifySignature(challenge, signature, pubKey) {
		logger.Warn("Signature verification failed")
		return false
	}

	return true
}

// readMessage reads a message from the connection.
// Format: [4-byte length][1-byte type][payload]
func (p *Peer) readMessage(conn net.Conn) (*Message, error) {
	logger := monitoring.GetLogger()
	reader := bufio.NewReader(conn)

	// Read message length (4 bytes)
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(reader, lengthBytes); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBytes)

	// Read message type (1 byte)
	typeByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	msgType := MessageType(typeByte)

	// Read payload
	payload := make([]byte, length-1) // Subtract 1 for type byte
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}

	// Decrypt payload
	decrypted, err := p.auth.Decrypt(payload)
	if err != nil {
		logger.Error("Failed to decrypt message", "error", err)
		return nil, err
	}

	return &Message{
		Type:    msgType,
		Payload: decrypted,
	}, nil
}

// processMessage handles the received message based on its type.
func (p *Peer) processMessage(msg *Message) error {
	logger := monitoring.GetLogger()

	switch msg.Type {
	case OrderRequest:
		// Parse order from payload (assuming format: ID|Type|Price|Quantity)
		parts := splitPayload(msg.Payload) // Custom function to split payload
		if len(parts) != 4 {
			return errors.New("invalid order request format")
		}

		orderType := trading.OrderType(parts[1])
		if orderType != trading.Buy && orderType != trading.Sell {
			return errors.New("invalid order type")
		}

		price, err := parseFloat(parts[2])
		if err != nil {
			return err
		}
		quantity, err := parseFloat(parts[3])
		if err != nil {
			return err
		}

		order := trading.NewOrder(parts[0], orderType, price, quantity)
		p.OrderBook.AddOrder(order)
		logger.Info("Order added", "id", order.ID, "type", order.Type)

		// Match orders and log trades
		trades := p.OrderBook.MatchOrders()
		for _, trade := range trades {
			logger.Info("Trade executed",
				"buyOrder", trade.BuyOrderID,
				"sellOrder", trade.SellOrderID,
				"price", trade.Price,
				"quantity", trade.Quantity)
		}

	case OrderConfirm:
		logger.Info("Order confirmation received", "payload", string(msg.Payload))

	case OrderCancel:
		logger.Info("Order cancellation received", "payload", string(msg.Payload))
		// Implement cancellation logic if needed

	default:
		return errors.New("unknown message type")
	}
	return nil
}

// Helper functions (not part of the original interface but needed)
func splitPayload(payload []byte) []string {
	// Simple split by '|' - in production, use a proper serialization format
	var parts []string
	start := 0
	for i, b := range payload {
		if b == '|' {
			parts = append(parts, string(payload[start:i]))
			start = i + 1
		}
	}
	parts = append(parts, string(payload[start:]))
	return parts
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
