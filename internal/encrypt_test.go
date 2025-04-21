package internal

import (
	"os"
	"testing"
)

func TestEncryptDecryptWithKey(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	original := "hello secret world"

	encrypted, err := EncryptWithKey([]byte(original), key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	decrypted, err := DecryptWithKey(encrypted, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("expected '%s', got '%s'", original, decrypted)
	}
}

func TestInitEncryptionKey(t *testing.T) {
	_ = os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = os.Setenv("TELEGRAM_ID_ENCRYPTION_KEY", "12345678901234567890123456789012")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "12345678901234567890123456789069")

	err := InitEncryptionKey()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(PayloadEncryptionKey) != 32 || len(SecretIDEncryptionKey) != 32 || len(SecretBotEncryptionKey) != 32 {
		t.Errorf("expected both keys to be 32 bytes, got: %d, %d, %d", len(PayloadEncryptionKey), len(SecretIDEncryptionKey), len(SecretBotEncryptionKey))
	}
}
