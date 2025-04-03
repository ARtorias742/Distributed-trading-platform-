package network

type MessageType int

const (
	OrderRequest MessageType = iota
	OrderConfirm
	OrderCancel
)

type Message struct {
	Type    MessageType
	Payload []byte
}
