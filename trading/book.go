package trading

import (
	"fmt"
	"sync"
	"time"
)

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

	// Iterate while there are buy and sell orders
	for len(ob.buyOrders) > 0 && len(ob.sellOrders) > 0 {
		// Get the first buy and sell orders
		buyOrder := ob.buyOrders[0]
		sellOrder := ob.sellOrders[0]

		// Check if the buy order price is greater than or equal to the sell order price
		if buyOrder.Price >= sellOrder.Price {
			// Determine the quantity to be match
			matchQuantity := min(buyOrder.Quantity, sellOrder.Quantity)

			// Update the quantities of the orders
			buyOrder.Quantity -= matchQuantity
			sellOrder.Quantity -= matchQuantity

			// Log or process the matched trade
			processTrade(buyOrder, sellOrder, matchQuantity)

			// Remove fully filled orders
			if buyOrder.Quantity == 0 {
				ob.buyOrders = ob.buyOrders[1:] // Remove the first order
			}

			if sellOrder.Quantity == 0 {
				ob.sellOrders = ob.sellOrders[1:] // Remove the first sell order
			}
		} else {
			// No more match
			break
		}
	}
}

func min(a float64, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

// Stub for processing a trade (now implemented)

func processTrade(buyOrder *Order, sellOrder *Order, quantity float64) {
	// Placeholder for processing a trade

	logTrade := Trade{
		BuyerID:   buyOrder.ID,
		SellerID:  sellOrder.ID,
		Quantity:  quantity,
		Price:     sellOrder.Price, // Trade happens at the sell price
		TradeTime: time.Now(),
	}

	// Add the trade to a trade history (if applicable)
	// Example: append to a global or OrderBook-specific trade history
	// ob.tradeHistory = append(ob.tradeHistory, logTrade)

	// Print the trade details (for debugging or logging purposes)
	fmt.Printf("Trade executed: Buyer %s, Seller %s, Quantity %d, Price %.2f\n",
		logTrade.BuyerID, logTrade.SellerID, logTrade.Quantity, logTrade.Price)
}
