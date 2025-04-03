package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
)

type AuthManager struct {
	key []byte
}

func NewAuthManager(key []byte) *AuthManager {
	return &AuthManager{key: key}
}

func (am *AuthManager) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(am.key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)
	return ciphertext, nil
}

func (am *AuthManager) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(am.key)
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}

// VerifySignature verifies an ECDSA signature for the given data using the provided public key.
// The signature is expected to be in the format [R || S], where R and S are concatenated big integers.
func VerifySignature(data []byte, signature []byte, pubKey []byte) bool {
	// Hash the data using SHA-256
	hash := sha256.Sum256(data)

	// parse the public key (assuming pubKey is in X509 encoded format)
	pub, err := ParseECDSAPublicKey(pubKey)
	if err != nil {
		return false
	}

	// Split the signature into R and S components (each typically 32 bytes for P-256 curve)
	if len(signature) != 64 { // 32 bytes for R + 32 bytes for S
		return false
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	//Verify the signature
	return ecdsa.Verify(pub, hash[:], r, s)
}

// ParseECDSAPublicKey is a helper function to parse an X509-encoded ECDSA public key.
// In a real system, this might be provided differently (e.g., from a PEM file).
func ParseECDSAPublicKey(pubKey []byte) (*ecdsa.PublicKey, error) {
	// This is a placeholder; in practice, you'd use x509.ParsePKIXPublicKey or similar
	// For this example, assume pubKey contains raw X and Y coordinates (64 bytes total)
	if len(pubKey) != 64 {
		return nil, io.ErrUnexpectedEOF
	}
	x := new(big.Int).SetBytes(pubKey[:32])
	y := new(big.Int).SetBytes(pubKey[32:])
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(), // Using P-256 curve; adjust as needed
		X:     x,
		Y:     y,
	}, nil
}
