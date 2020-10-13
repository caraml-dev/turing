package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
)

// CryptoService is used for encrypting / decrypting sensitive data
type CryptoService interface {
	// Encrypt takes an input plaintext string and returs the cipher text or an error
	Encrypt(plaintext string) (string, error)
	// Decrypt takes an input cipher string and returs the plaintext or an error
	Decrypt(ciphertext string) (string, error)
}

type cryptoService struct {
	encryptionKey string
}

// NewCryptoService creates a new cryptoService using the given encryption key
func NewCryptoService(encryptionKey string) CryptoService {
	return &cryptoService{
		encryptionKey: createHash(encryptionKey),
	}
}

func (cs *cryptoService) Encrypt(plainText string) (string, error) {
	key, err := hex.DecodeString(cs.encryptionKey)
	if err != nil {
		return "", err
	}
	block, _ := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (cs *cryptoService) Decrypt(cipherText string) (string, error) {
	key, err := hex.DecodeString(cs.encryptionKey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	ciphertextByte, _ := base64.StdEncoding.DecodeString(cipherText)
	nonce, ciphertext := ciphertextByte[:nonceSize], ciphertextByte[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func createHash(key string) string {
	hasher := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hasher[:])
}
