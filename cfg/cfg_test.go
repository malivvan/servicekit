package cfg

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type TestConfig struct {
	Hostname string
	Port     int
	Password string `encrypt:"obfuscaaaaate!"`
}

func TestLoad(t *testing.T) {
	defer os.Remove("config.yaml")

	x1 := TestConfig{
		Hostname: "hello",
		Port:     8080,
		Password: "dontlooook!",
	}
	x2 := TestConfig{}

	err := LoadFile("config.yaml", &x1)
	assert.NoError(t, err)

	err = LoadFile("config.yaml", &x2)
	assert.NoError(t, err)

	assert.Equal(t, x1, x2)
}
