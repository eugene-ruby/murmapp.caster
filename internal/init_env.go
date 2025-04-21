package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"fmt"
)

var SecretSalt string
var WebhookHost string
var TelegramAPI string

func InitEnv() error {
	TelegramAPI = os.Getenv("TELEGRAM_API_URL")
	if TelegramAPI == "" {
		return fmt.Errorf("TELEGRAM_API_URL env var not set")
	}

   SecretSalt = os.Getenv("SECRET_SALT")
	if SecretSalt == "" || len(SecretSalt) < 8 {
		return fmt.Errorf("SECRET_SALT env var not set")
	}

	WebhookHost = os.Getenv("WEB_HOOK_HOST")
	if WebhookHost == "" {
		return fmt.Errorf("WEB_HOOK_HOST env var not set")
	}

	return nil
}
