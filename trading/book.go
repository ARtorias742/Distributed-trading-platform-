package trading

import (
	"sort"
	"sync"
	"time"

	"github.com/artorias742/DTP/monitoring"
)

type Trade struct {
	BuyOrderID  string
	SellOrderID string
	Price       float64
	Quantity    float64
}

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

func (ob *OrderBook) MatchOrders() []Trade {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	start := time.Now()
	defer func() {
		monitoring.OrderLatency.Observe(time.Since(start).Seconds())
	}()

	var trades []Trade
	sort.Slice(ob.buyOrders, func(i, j int) bool {
		return ob.buyOrders[i].Price > ob.buyOrders[j].Price ||
			(ob.buyOrders[i].Price == ob.buyOrders[j].Price &&
				ob.buyOrders[i].Timestamp.Before(ob.buyOrders[j].Timestamp))
	})
	sort.Slice(ob.sellOrders, func(i, j int) bool {
		return ob.sellOrders[i].Price < ob.sellOrders[j].Price ||
			(ob.sellOrders[i].Price == ob.sellOrders[j].Price &&
				ob.sellOrders[i].Timestamp.Before(ob.sellOrders[j].Timestamp))
	})

	for len(ob.buyOrders) > 0 && len(ob.sellOrders) > 0 {
		buy := ob.buyOrders[0]
		sell := ob.sellOrders[0]
		if buy.Price >= sell.Price {
			quantity := min(buy.Quantity, sell.Quantity)
			trade := Trade{
				BuyOrderID:  buy.ID,
				SellOrderID: sell.ID,
				Price:       sell.Price,
				Quantity:    quantity,
			}
			trades = append(trades, trade)

			buy.Quantity -= quantity
			sell.Quantity -= quantity
			if buy.Quantity == 0 {
				ob.buyOrders = ob.buyOrders[1:]
			}
			if sell.Quantity == 0 {
				ob.sellOrders = ob.sellOrders[1:]
			}
		} else {
			break
		}
	}
	return trades
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
