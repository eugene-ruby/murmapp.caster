package internal_test

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"murmapp.caster/internal"
)

func TestInitEncryptionKey_InvalidLength(t *testing.T) {
	_ = os.Setenv("ENCRYPTION_KEY", "short")
	_ = os.Setenv("TELEGRAM_ID_ENCRYPTION_KEY", "stillshort")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "123")

	err := internal.InitEncryptionKey()
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be 32 bytes")
}

func TestDecryptWithKey_InvalidBase64(t *testing.T) {
	badData := []byte("this is not encrypted")
	key := []byte("01234567890123456789012345678901")

	_, err := internal.DecryptWithKey(badData, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decryption failed")
}

func TestDecryptWithKey_CorruptedCiphertext(t *testing.T) {
	// base64 but not valid AES-GCM
	bad, _ := base64.URLEncoding.DecodeString("dGVzdCBzdHJpbmc=")
	key := []byte("01234567890123456789012345678901")

	_, err := internal.DecryptWithKey(bad, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ciphertext too short")
}
