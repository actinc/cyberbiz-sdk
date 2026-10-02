package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	c, err := New(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)
	sealed, err := c.Encrypt([]byte("secret"))
	require.NoError(t, err)
	assert.NotContains(t, string(sealed), "secret")
	plain, err := c.Decrypt(sealed)
	require.NoError(t, err)
	assert.Equal(t, "secret", string(plain))

	empty, err := c.Encrypt(nil)
	require.NoError(t, err)
	assert.Nil(t, empty)
	_, err = New([]byte("short"))
	assert.Error(t, err)
	sealed[len(sealed)-1] ^= 1
	_, err = c.Decrypt(sealed)
	assert.Error(t, err)
}

func TestFingerprint(t *testing.T) {
	assert.Regexp(t, `^sha256:[0-9a-f]{8} len=5$`, Fingerprint("token"))
	assert.Equal(t, "", Fingerprint(""))
}
