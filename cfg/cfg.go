package cfg

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/malivvan/servicekit"
)

const (
	cryptoPrefix = "$("
	cryptoSuffix = ")"
	cryptoTag    = "encrypt"
)

func LoadFile(path string, config interface{}) error {
	if path == "" {
		path = servicekit.Workdir(servicekit.Name() + ".yaml")
	}

	// 1. read bytes, if not exist write initial config
	configBytes, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		configBytes, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, configBytes, 0600)
		if err != nil {
			return err
		}
	}

	// 2. decode config and encrypt all unencrypted strings
	configBytes, err = os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return err
	}
	if err := alterStrings(config, func(data, secret string) (string, error) {
		if !(strings.HasPrefix(data, cryptoPrefix) && strings.HasSuffix(data, cryptoSuffix)) {
			encrypted, err := encrypt(data, secret)
			if err != nil {
				return "", err
			}
			return cryptoPrefix + base64.StdEncoding.EncodeToString([]byte(encrypted)) + cryptoSuffix, nil
		}
		return data, nil
	}); err != nil {
		return err
	}

	// 3. if config was altered save to disk
	newConfigBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if !bytes.Equal(newConfigBytes, configBytes) {
		err := os.WriteFile(path, newConfigBytes, 0600)
		if err != nil {
			return err
		}
	}

	// 4. decrypt all strings
	if err := alterStrings(config, func(data, secret string) (string, error) {
		if strings.HasPrefix(data, cryptoPrefix) && strings.HasSuffix(data, cryptoSuffix) {
			encryptedData, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(strings.TrimPrefix(data, cryptoPrefix), cryptoSuffix))
			if err != nil {
				return "", err
			}
			return decrypt(string(encryptedData), secret)
		}
		return data, nil
	}); err != nil {
		return err
	}

	return nil
}

func alterStrings(v interface{}, f func(data, secret string) (string, error)) error {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			for field.Kind() == reflect.Ptr {
				field = field.Elem()
			}
			switch field.Kind() {
			case reflect.Struct:
				err := alterStrings(val.Field(i).Addr().Interface(), f)
				if err != nil {
					return err
				}
			case reflect.String:
				if _, ok := val.Type().Field(i).Tag.Lookup(cryptoTag); ok {
					tagValue := val.Type().Field(i).Tag.Get(cryptoTag)
					str, err := f(field.String(), tagValue)
					if err != nil {
						return err
					}
					field.SetString(str)
				}
			}
		}
	}
	return nil
}

func encrypt(plaintext, secretKey string) (string, error) {
	for i := len(secretKey); i < 32; i++ {
		secretKey += strconv.Itoa(i)
	}
	secretKey = secretKey[:32]
	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return string(ciphertext), nil
}

func decrypt(ciphertext, secretKey string) (string, error) {
	for i := len(secretKey); i < 32; i++ {
		secretKey += strconv.Itoa(i)
	}
	secretKey = secretKey[:32]
	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
