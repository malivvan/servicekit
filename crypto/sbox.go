package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	sboxKeyLen   = 32
	sboxNonceLen = 24
)

func EncryptSbox(key []byte, data []byte) ([]byte, error) {
	if len(key) != sboxKeyLen {
		return nil, errors.New("wrong secretbox key length")
	}
	nonce, err := generateSBoxNonce()
	if err != nil {
		return nil, errors.New("encryption failed")
	}
	out := make([]byte, len(nonce))
	copy(out, nonce[:])
	out = secretbox.Seal(out, data, nonce, toSboxKey(key))
	return out, nil
}

// panics if key is nil
func DecryptSbox(key []byte, data []byte) ([]byte, error) {
	if len(key) != sboxKeyLen {
		return nil, errors.New("wrong secretbox key length")
	}
	if len(data) < (sboxNonceLen + secretbox.Overhead) {
		return nil, errors.New("decryption failed")
	}
	var nonce [sboxNonceLen]byte
	copy(nonce[:], data[:sboxNonceLen])
	out, ok := secretbox.Open(nil, data[sboxNonceLen:], &nonce, toSboxKey(key))
	if !ok {
		return nil, errors.New("decryption failed")
	}
	return out, nil
}

func generateSBoxNonce() (*[sboxNonceLen]byte, error) {
	nonce := new([sboxNonceLen]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func toSboxKey(b []byte) *[sboxKeyLen]byte {
	var key [sboxKeyLen]byte
	for i := range key {
		key[i] = b[i]
	}
	return &key
}
