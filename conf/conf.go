package conf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/malivvan/servicekit"
	"github.com/malivvan/servicekit/crypto"
)

const (
	cryptoPrefix = "$("
	cryptoSuffix = ")"
	cryptoTag    = "encrypt"
)

func Load(secret string, config interface{}) error {
	path := servicekit.Workdir(servicekit.Name() + ".json")

	// 1. read bytes, if not exist write initial config
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		configBytes, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path, configBytes, 0600)
		if err != nil {
			return err
		}
	}

	// 2. decode config and encrypt all unencrypted strings
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return err
	}
	if err := alterStrings(config, func(str string) (string, error) {
		if !(strings.HasPrefix(str, cryptoPrefix) && strings.HasSuffix(str, cryptoSuffix)) {
			encryptedData, err := crypto.Encrypt(nil, secret, []byte(str))
			if err != nil {
				return "", err
			}
			return cryptoPrefix + base64.StdEncoding.EncodeToString(encryptedData) + cryptoSuffix, nil
		}
		return str, nil
	}); err != nil {
		return err
	}

	// 3. if config was altered save to disk
	newConfigBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if !bytes.Equal(newConfigBytes, configBytes) {
		err = ioutil.WriteFile(path, newConfigBytes, 0600)
		if err != nil {
			return err
		}
	}

	// 4. decrypt all strings
	if err := alterStrings(config, func(str string) (string, error) {
		if strings.HasPrefix(str, cryptoPrefix) && strings.HasSuffix(str, cryptoSuffix) {
			encryptedData, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(strings.TrimPrefix(str, cryptoPrefix), cryptoSuffix))
			if err != nil {
				return "", err
			}
			decryptedData, _, err := crypto.Decrypt(secret, encryptedData)
			if err != nil {
				return "", err
			}
			return string(decryptedData), nil
		}
		return str, nil
	}); err != nil {
		return err
	}

	return nil
}

func alterStrings(v interface{}, f func(str string) (string, error)) error {
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
					str, err := f(field.String())
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
