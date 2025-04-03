package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/artorias742/DTP/monitoring"
	"github.com/artorias742/DTP/network"
	"github.com/artorias742/DTP/trading"
	"github.com/google/uuid"
)

type Server struct {
	peer   *network.Peer
	mutex  sync.Mutex
	orders chan *trading.Order // Channel to queue orders for processing
}

func NewServer(peer *network.Peer) *Server {
	return &Server{
		peer:   peer,
		orders: make(chan *trading.Order, 100), // Buffered channel for up to 100 orders
	}
}

// Start begins the HTTP server and order processing goroutine.
func (s *Server) Start() {
	logger := monitoring.GetLogger()

	// Start order processing in the background
	go s.processOrders()

	// Define HTTP endpoints
	http.HandleFunc("/order", s.handleOrder)
	http.HandleFunc("/health", s.handleHealth)

	// Start server on port :8083
	logger.Info("Starting API server", "addr", ":8083")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		logger.Error("API server failed", "error", err)
	}
}

// handleOrder handles POST requests to place a new order.
func (s *Server) handleOrder(w http.ResponseWriter, r *http.Request) {
	logger := monitoring.GetLogger()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type     string  `json:"type"`
		Price    float64 `json:"price"`
		Quantity float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		logger.Error("Failed to decode order request", "error", err)
		return
	}

	// Validate order type
	orderType := trading.OrderType(req.Type)
	if orderType != trading.Buy && orderType != trading.Sell {
		http.Error(w, "Invalid order type", http.StatusBadRequest)
		return
	}

	// Create and queue order
	order := trading.NewOrder(uuid.New().String(), orderType, req.Price, req.Quantity)
	s.orders <- order // Send to processing channel

	logger.Info("Order received from user",
		"id", order.ID,
		"type", order.Type,
		"price", order.Price,
		"quantity", order.Quantity)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"order_id": order.ID})
}

// handleHealth provides a simple health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// processOrders processes orders from the channel and adds them to the order book.
func (s *Server) processOrders() {
	logger := monitoring.GetLogger()
	for order := range s.orders {
		// Add order to the order book
		s.addOrder(order)

		// Match orders after adding
		trades := s.peer.OrderBook.MatchOrders()

		logger.Info("Order processed", "id", order.ID)
		for _, trade := range trades {
			logger.Info("Trade executed",
				"buyOrder", trade.BuyOrderID,
				"sellOrder", trade.SellOrderID,
				"price", trade.Price,
				"quantity", trade.Quantity)
		}
	}
}

// addOrder adds an order to the peer's order book with proper synchronization.
func (s *Server) addOrder(order *trading.Order) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Directly call AddOrder from trading/book.go
	s.peer.OrderBook.AddOrder(order)
	logger := monitoring.GetLogger()
	logger.Debug("Order added to book",
		"id", order.ID,
		"type", order.Type,
		"price", order.Price,
		"quantity", order.Quantity)
}
