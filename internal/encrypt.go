package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

var SecretIDEncryptionKey []byte  // for encryption telegram_id, know hook and caster
var SecretBotEncryptionKey []byte // for encryption bot_api_key, know only caster
var PayloadEncryptionKey []byte   // for encryption payload of message, know hook, caster, core

func InitEncryptionKey() error {
	payloadKey := os.Getenv("ENCRYPTION_KEY")
	if payloadKey == "" || len(payloadKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be 32 bytes")
	}
	PayloadEncryptionKey = []byte(payloadKey)

	encIDKey := os.Getenv("TELEGRAM_ID_ENCRYPTION_KEY")
	if encIDKey == "" || len(encIDKey) != 32 {
		return fmt.Errorf("TELEGRAM_ID_ENCRYPTION_KEY must be 32 bytes")
	}
	SecretIDEncryptionKey = []byte(encIDKey)

	encBotKey := os.Getenv("BOT_ENCRYPTION_KEY")
	if encBotKey == "" || len(encBotKey) != 32 {
		return fmt.Errorf("BOT_ENCRYPTION_KEY must be 32 bytes")
	}
	SecretBotEncryptionKey = []byte(encBotKey)

	return nil
}

func EncryptWithKey(plain []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}

func DecryptWithKey(ciphertext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher init failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM init failed: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	data := ciphertext[gcm.NonceSize():]

	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plain), nil
}
