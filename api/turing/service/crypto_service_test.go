package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoService(t *testing.T) {
	testphrase := "test"
	cs := NewCryptoService("key")
	cipher, err := cs.Encrypt(testphrase)
	assert.NoError(t, err)
	assert.NotEqual(t, testphrase, cipher)
	plaintext, err := cs.Decrypt(cipher)
	assert.NoError(t, err)
	assert.Equal(t, testphrase, plaintext)
}
