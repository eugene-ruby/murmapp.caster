package registration

import (
	"math/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"bytes"
	"encoding/json"
	"net/http"
)

func ComputeWebhookID(secretToken, secretSalt string) string {
	h := sha256.New()
	h.Write([]byte(secretToken + secretSalt))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateSecretToken() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RegisterTelegramWebhook(apiKey, webhookURL, secretToken, telegramAPI string) error {
	apiURL := fmt.Sprintf("%s/bot%s/setWebhook", telegramAPI, apiKey)

	payload := map[string]string{
		"url":          webhookURL,
		"secret_token": secretToken,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
