package mon

import (
	"crypto/sha1"
	"encoding/base32"
)

func OSID() (string, error) {
	id, err := getOSID()
	if err != nil {
		return "", err
	}
	hasher := sha1.New()
	hasher.Write([]byte{207, 17, 194, 72, 157, 77, 8, 86, 224, 25, 62, 137, 69, 48, 27, 174})
	hasher.Write([]byte(id))
	hasher.Write([]byte{36, 107, 113, 108, 118, 186, 94, 181, 187, 20, 108, 247, 159, 59, 166, 54})
	return base32.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}
