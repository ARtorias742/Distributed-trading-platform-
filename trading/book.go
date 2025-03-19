package trading

import "sync"

type OrderBook struct {
	buyOrders  []*Order
	sellOrders []*Order
	mutex      sync.Mutex
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyOrders:  make([]*Order, 0),
		sellOrders: make([]*Order, 0),
	}
}


func (ob *OrderBook) AddOrder(order *Order) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	if order.Type == Buy {
		ob.buyOrders = append(ob.buyOrders, order)

	} else {
		ob.sellOrders = append(ob.sellOrders, order)
	}
}


func (ob *OrderBook) MatchOrders() {
	//
	ob.mutex.Lock()
	defer ob.mutex.Unlock()
	
}