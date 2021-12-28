package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScryptConfig_EncodeDecode(t *testing.T) {
	encodedConfig := DefaultScryptConfig.Encode()
	config, err := DecodeScryptConfig(encodedConfig)
	assert.NoError(t, err)

	assert.Equal(t, DefaultScryptConfig.N, config.N)
	assert.Equal(t, DefaultScryptConfig.R, config.R)
	assert.Equal(t, DefaultScryptConfig.P, config.P)
}
