package network

import (
	"log"
	"net"

	"github.com/ARtorias742/DTP/config"
	"github.com/ARtorias742/DTP/trading"
)

type Peer struct {
	config    *config.Config
	orderBook *trading.OrderBook
	listener  net.Listener
	peers     map[string]net.Conn
}

func NewPeer(cfg *config.Config) *Peer {
	return &Peer{
		config:    cfg,
		orderBook: trading.NewOrderBook(),
		peers:     make(map[string]net.Conn),
	}
}

func (p *Peer) Start() error {
	listener, err := net.Listen("tcp", p.config.ListenAddr)
	if err != nil {
		return err
	}

	p.listener = listener

	// start accepting connections
	go p.acceptConnections()

	// Connect to seed nodes
	p.connectToSeeds()

	log.Printf("Peer started on %s", p.config.ListenAddr)
	return nil
}

func (p *Peer) acceptConnections() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err)
			continue
		}

		go p.handleConnection(conn)
	}
}

func (p *Peer) handleConnection(conn net.Conn) {
	defer func() {
		// Remove the connection from peers and close it when the function exits
		delete(p.peers, conn.RemoteAddr().String())
		conn.Close()
		log.Printf("Connection closed: %s", conn.RemoteAddr().String())
	}()

	log.Printf("New connection established: %s", conn.RemoteAddr().String())

	// Add connection to peers
	p.peers[conn.RemoteAddr().String()] = conn

	// Buffer to read incoming data
	buffer := make([]byte, 1024)

	for {
		// Read data from the connection
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				log.Printf("Connection closed by peer: %s", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading from connection %s: %s", conn.RemoteAddr().String(), err)
			}
			break
		}

		// Process the received message
		message := string(buffer[:n])
		log.Printf("Received message from %s: %s", conn.RemoteAddr().String(), message)

		// TODO: Add logic to handle the message (e.g., parse and act on it)
	}
}

func (p *Peer) connectToSeeds() {
	for _, addr := range p.config.SeedNodes {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("Failed to connect to seed node %s: %s", addr, err)
			continue
		}
		p.peers[addr] = conn
	}
}
