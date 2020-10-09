package mocks

import "github.com/stretchr/testify/mock"

// CryptoService implements CryptoService interface
type CryptoService struct {
	mock.Mock
}

// Encrypt implements CryptoService.Encrypt
func (cs *CryptoService) Encrypt(text string) (string, error) {
	ret := cs.Called(text)

	var err error
	if ret[1] != nil {
		err = (ret[1]).(error)
	}

	return (ret[0]).(string), err
}

// Decrypt implements CryptoService.Decrypt
func (cs *CryptoService) Decrypt(cipher string) (string, error) {
	ret := cs.Called(cipher)

	var err error
	if ret[1] != nil {
		err = (ret[1]).(error)
	}

	return (ret[0]).(string), err
}
