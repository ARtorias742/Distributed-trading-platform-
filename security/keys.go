package security

type KeyManager struct {
	privateKey []byte
	publicKey  []byte
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		privateKey: []byte("placeholder-private-key"),
		publicKey:  []byte("placeholder-public-key"),
	}
}
