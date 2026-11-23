package mongo_kit

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey      = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	otherTestKey = "1f1e1d1c1b1a191817161514131211100f0e0d0c0b0a09080706050403020100"
)

func TestFieldCipher_RoundTrip(t *testing.T) {
	c, err := newFieldCipher(testKey)
	require.NoError(t, err)

	for _, plaintext := range []string{"", "secret", "ñandú 🔐", strings.Repeat("x", 10_000)} {
		enc, err := c.encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, enc)

		dec, err := c.decrypt(enc)
		require.NoError(t, err)
		assert.Equal(t, plaintext, dec)
	}
}

func TestFieldCipher_RandomNonce(t *testing.T) {
	c, err := newFieldCipher(testKey)
	require.NoError(t, err)

	a, err := c.encrypt("same")
	require.NoError(t, err)
	b, err := c.encrypt("same")
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "each encryption uses a fresh nonce")
}

func TestFieldCipher_DecryptErrors(t *testing.T) {
	c, err := newFieldCipher(testKey)
	require.NoError(t, err)
	other, err := newFieldCipher(otherTestKey)
	require.NoError(t, err)

	enc, err := other.encrypt("secret")
	require.NoError(t, err)

	tests := map[string]string{
		"other key":  enc,
		"not base64": "%%%",
		"too short":  base64.StdEncoding.EncodeToString([]byte("abc")),
		"tampered":   enc[:len(enc)-4] + "AAAA",
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := c.decrypt(value)
			assert.ErrorIs(t, err, ErrDecryptionFailed)
		})
	}
}

// Values encrypted before this library had encryption (Guardian's encryptor:
// AES-256-GCM, base64 of nonce + ciphertext + tag) must stay readable.
func TestFieldCipher_CompatibleFormat(t *testing.T) {
	key, _ := hex.DecodeString(testKey)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	legacy := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte("legacy token"), nil))

	c, err := newFieldCipher(testKey)
	require.NoError(t, err)
	dec, err := c.decrypt(legacy)
	require.NoError(t, err)
	assert.Equal(t, "legacy token", dec)
}

func TestNewFieldCipher_InvalidKey(t *testing.T) {
	for _, key := range []string{"", "zz", testKey[:62], testKey + "00"} {
		_, err := newFieldCipher(key)
		assert.Error(t, err, "key %q", key)
	}
}

func TestClient_EncryptionDisabled(t *testing.T) {
	c := &Client{}
	assert.False(t, c.EncryptionEnabled())

	_, err := c.Encrypt("x")
	assert.ErrorIs(t, err, ErrEncryptionDisabled)
	_, err = c.Decrypt("x")
	assert.ErrorIs(t, err, ErrEncryptionDisabled)
}

func TestClient_EncryptDecrypt(t *testing.T) {
	fc, err := newFieldCipher(testKey)
	require.NoError(t, err)
	c := &Client{cipher: fc}
	assert.True(t, c.EncryptionEnabled())

	enc, err := c.Encrypt("token")
	require.NoError(t, err)
	dec, err := c.Decrypt(enc)
	require.NoError(t, err)
	assert.Equal(t, "token", dec)
}

func TestConfigValidate_EncryptionKey(t *testing.T) {
	cfg := DefaultConfig()

	WithEncryptionKey(testKey)(&cfg)
	assert.NoError(t, cfg.validate())

	for _, key := range []string{"change-me-in-production", testKey[:62], testKey + "00"} {
		WithEncryptionKey(key)(&cfg)
		err := cfg.validate()
		var cfgErr *ConfigError
		require.True(t, errors.As(err, &cfgErr), "key %q", key)
		assert.Equal(t, "EncryptionKey", cfgErr.Field)
	}
}

func TestResolveDatabase(t *testing.T) {
	tests := []struct {
		name, uri, database, want string
	}{
		{"from URI path", "mongodb://u:p@host:27017/guardian_staging?authSource=guardian_staging", "", "guardian_staging"},
		{"explicit wins", "mongodb://u:p@host:27017/other", "mine", "mine"},
		{"no database anywhere", "mongodb://host:27017/?authSource=admin", "", ""},
		{"escaped password", "mongodb://u:p%40ss@host:27017/app", "", "app"},
		{"invalid URI left alone", "not-a-uri", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{URI: tt.uri, Database: tt.database}
			cfg.resolveDatabase()
			assert.Equal(t, tt.want, cfg.Database)
		})
	}
}

// New validates the final configuration (after the options), before connecting.
func TestNew_ValidatesOptions(t *testing.T) {
	_, err := New(DefaultConfig(), WithEncryptionKey("not-hex"))
	var cfgErr *ConfigError
	require.True(t, errors.As(err, &cfgErr))
	assert.Equal(t, "EncryptionKey", cfgErr.Field)

	_, err = New(DefaultConfig(), WithURI("mongodb://localhost:27017/?authSource=admin"), WithDatabase(""))
	require.True(t, errors.As(err, &cfgErr))
	assert.Equal(t, "Database", cfgErr.Field)
}
