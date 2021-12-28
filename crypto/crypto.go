// package crypto provides helper functions for an established set of crypto libraries (nacl, scrypt)
package crypto

import (
	"errors"

	"golang.org/x/crypto/nacl/secretbox"
)

// Encrypts given data with nacl/secretbox using a scrypt derived key. The first 24 bytes of the resulting byte slice
// contain the scrypt configuration. The following 32 bytes contain the scrypt salt. All remaining bytes represent the
// encrypted data.
func Encrypt(scryptConfig *ScryptConfig, password string, data []byte) ([]byte, error) {

	if scryptConfig == nil {
		scryptConfig = &DefaultScryptConfig
	}

	salt, err := GenerateScryptSalt()
	if err != nil {
		return nil, err
	}

	key, err := scryptConfig.Derive(salt, sboxKeyLen, password)
	if err != nil {
		return nil, err
	}

	encryptedData, err := EncryptSbox(key, data)
	if err != nil {
		return nil, err
	}

	return append(scryptConfig.Encode(), append(salt, encryptedData...)...), nil
}

func Decrypt(password string, data []byte) ([]byte, *ScryptConfig, error) {
	if len(data) < (scryptConfigLen + scryptSaltLen + sboxNonceLen + secretbox.Overhead) {
		return nil, nil, errors.New("decryption failed")
	}

	config := data[:scryptConfigLen]
	salt := data[scryptConfigLen : scryptConfigLen+scryptSaltLen]
	data = data[scryptConfigLen+scryptSaltLen:]

	scryptConfig, err := DecodeScryptConfig(config)
	if err != nil {
		return nil, nil, err
	}

	key, err := scryptConfig.Derive(salt, sboxKeyLen, password)
	if err != nil {
		return nil, nil, err
	}

	decryptedData, err := DecryptSbox(key, data)
	if err != nil {
		return nil, nil, err
	}

	return decryptedData, &scryptConfig, err
}
