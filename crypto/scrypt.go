package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptSaltLen   = 32
	scryptConfigLen = 24
)

type ScryptConfig struct {
	// CPU/memory cost parameter (logN)
	N uint64 `json:"n"`

	// block size parameter (octets)
	R uint64 `json:"r"`

	// parallelisation parameter (positive int)
	P uint64 `json:"p"`
}

// requires 32MB RAM for under a second (i7).
var DefaultScryptConfig = ScryptConfig{
	N: 1 << uint64(15),
	R: 8,
	P: 1,
}

func DecodeScryptConfig(data []byte) (ScryptConfig, error) {
	if len(data) != scryptConfigLen {
		return ScryptConfig{}, errors.New("wrong scrypt config length")
	}
	return ScryptConfig{
		N: binary.LittleEndian.Uint64(data[0:8]),
		R: binary.LittleEndian.Uint64(data[8:16]),
		P: binary.LittleEndian.Uint64(data[16:24]),
	}, nil
}

func (config ScryptConfig) Encode() []byte {
	var data [scryptConfigLen]byte
	binary.LittleEndian.PutUint64(data[0:8], config.N)
	binary.LittleEndian.PutUint64(data[8:16], config.R)
	binary.LittleEndian.PutUint64(data[16:24], config.P)
	return data[:]
}

func (config ScryptConfig) MemoryRequiredMB() int {
	return (int(config.N) * int(config.R) * 128) / 1024 / 1024
}

func (config ScryptConfig) TimeRequiredMS() (int, error) {
	t := time.Now()
	salt, err := GenerateScryptSalt()
	if err != nil {
		return -1, err
	}
	_, err = config.Derive(salt, 32, "selftest")
	if err != nil {
		return -1, err
	}
	return int(time.Since(t) / time.Millisecond), nil
}

func GenerateScryptSalt() ([]byte, error) {
	salt := make([]byte, scryptSaltLen)
	_, err := io.ReadFull(rand.Reader, salt[:])
	if err != nil {
		return nil, errors.New("error creating salt")
	}
	return salt[:], nil
}

func (config ScryptConfig) Derive(salt []byte, keyLen int, password string) ([]byte, error) {
	b, err := scrypt.Key([]byte(password), salt, int(config.N), int(config.R), int(config.P), keyLen)
	if err != nil {
		return nil, err
	}
	if len(b) != keyLen {
		return nil, errors.New("derived key has wrong length")
	}
	return b, nil
}
