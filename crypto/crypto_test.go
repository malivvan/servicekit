package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {

	var message [256]byte
	_, err := io.ReadFull(rand.Reader, message[:])
	if err != nil {
		t.Fatal(err)
	}
	enc, err := Encrypt(nil, "secret", message[:])
	if err != nil {
		t.Fatal(err)
	}
	dec, _, err := Decrypt("secret", enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(message[:], dec) {
		t.Fatal("decrypted message does not match original message")
	}
}
