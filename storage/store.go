package storage

import (
	"sync"

	"github.com/ARtorias742/DTP/trading"
)

type Store struct {
	orders map[string]*trading.Order
	mutex  sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		orders: make(map[string]*trading.Order),
	}
}

func (s *Store) SaveOrder(order *trading.Order) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.orders[order.ID] = order
}

func (s *Store) GetOrder(id string) *trading.Order {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.orders[id]
}
