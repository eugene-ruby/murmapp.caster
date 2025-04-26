package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
)

var CachedPrivateKey *rsa.PrivateKey

func InitPrivateRSA() error {
	encPrivateKey := os.Getenv("ENCRYPTED_PRIVATE_KEY")
	key := deriveKey([]byte(MasterEncryptionKey), "CASTER_PRIVATE_RSA")
	err := fetchPrivateRSA(encPrivateKey, key)
	if err != nil {
		return fmt.Errorf("fetch private key failed: %w", err)
	}
	return nil
}

func deriveKey(master []byte, label string) []byte {
	h := hmac.New(sha256.New, master)
	h.Write([]byte(label))
	return h.Sum(nil)[:32]
}

func fetchPrivateRSA(encryptedBase64 string, masterKey []byte) error {
	if CachedPrivateKey != nil {
		return nil
	}

	ciphertext, err := base64.RawURLEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return fmt.Errorf("base64 decode failed: %w", err)
	}

	plaintext, err := DecryptWithKey(ciphertext, masterKey)
	if err != nil {
		return fmt.Errorf("decrypt failed: %w", err)
	}

	pemBlock, _ := pem.Decode([]byte(plaintext))
	if pemBlock == nil {
		return errors.New("empty private key")
	} else if pemBlock.Type != "RSA PRIVATE KEY" {
		return errors.New("invalid PEM block for private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse private key failed: %w", err)
	}

	CachedPrivateKey = priv
	return nil
}
