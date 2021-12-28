package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestSboxEncryptDecrypt(t *testing.T) {
	var secret [32]byte
	_, err := io.ReadFull(rand.Reader, secret[:])
	if err != nil {
		t.Fatal(err)
	}
	var message [256]byte
	_, err = io.ReadFull(rand.Reader, message[:])
	if err != nil {
		t.Fatal(err)
	}
	enc, err := EncryptSbox(secret[:], message[:])
	if err != nil {
		t.Fatal(err)
	}
	dec, err := DecryptSbox(secret[:], enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(message[:], dec) {
		t.Fatal("decrypted message does not match original message")
	}
}
