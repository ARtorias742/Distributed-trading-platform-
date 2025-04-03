package trading

import "time"

type OrderType string

const (
	Buy  OrderType = "BUY"
	Sell OrderType = "SELL"
)

type Order struct {
	ID        string
	Type      OrderType
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

func NewOrder(id string, orderType OrderType, price, quantity float64) *Order {
	return &Order{
		ID:        id,
		Type:      orderType,
		Price:     price,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}
}
