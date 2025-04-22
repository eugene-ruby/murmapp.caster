package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

var HandlerMessageOut = func(body []byte, apiBase string) {
	var req casterpb.SendMessageRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		log.Printf("❌ Failed to decode proto: %v", err)
		return
	}

	decryptApiKey, err := DecryptWithKey(req.EncryptedApiKeyBot, SecretBotEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt bot api key: %v", err)
		return
	}

	decryptPayload, err := DecryptWithKey(req.EncryptedPayload, PayloadEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt bot api key: %v", err)
		return
	}
	if !json.Valid([]byte(decryptPayload)) {
		log.Printf("⚠️ Decrypted payload is not valid JSON!")
		return
	}

	url := fmt.Sprintf("%s/bot%s/%s", apiBase, decryptApiKey, req.ApiEndpoint)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader([]byte(decryptPayload)))
	if err != nil {
		log.Printf("🚫 Failed to create HTTP request: %v", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("📡 Failed to send to Telegram API: %v", err)
		return
	}
	defer resp.Body.Close()

	safeURL := fmt.Sprintf("%s/bot[redacted]/%s", apiBase, req.ApiEndpoint)
	log.Printf("✅ Telegram API response: %s | → %s", resp.Status, safeURL)
}
