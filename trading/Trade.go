package trading

import "time"

type Trade struct {
	BuyerID   string
	SellerID  string
	Quantity  float64
	Price     float64
	TradeTime time.Time
}
