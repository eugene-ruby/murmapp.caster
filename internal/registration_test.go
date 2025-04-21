package internal

import (
	"encoding/base64"
	"os"
	"testing"
	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

func TestHandleRegistrationMessage(t *testing.T) {
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = InitEncryptionKey()

	apiKey := "987654:AAA-abcXYZ"
	botID := "testbot"

	encrypted, err := EncryptWithKey(apiKey, SecretBotEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt apiKey: %v", err)
	}
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("failed to decode encrypted apiKey: %v", err)
	}

	req := &casterpb.RegisterWebhookRequest{
		BotId:      botID,
		ApiKeyBot:  ciphertext,
	}
	payload, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal proto: %v", err)
	}

	// Just ensure handler does not panic
	handleRegistrationMessage(payload, "https://api.telegram.org")
}

func TestHandleMessageOut_ValidPayload(t *testing.T) {
	_ = os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = InitEncryptionKey()

	json := `{"chat_id": 12345, "text": "hello world"}`
	payload, err := EncryptWithKeyBytes([]byte(json), PayloadEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt payload: %v", err)
	}

	apiKey := "987654:AAA-abcXYZ"
	encrypted, err := EncryptWithKey(apiKey, SecretBotEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt api key: %v", err)
	}
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}

	req := &casterpb.SendMessageRequest{
		ApiEndpoint:      "sendMessage",
		EncryptedApiKeyBot: ciphertext,
		EncryptedPayload:   payload,
	}
	msg, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("proto marshal failed: %v", err)
	}

	// Test should not panic
	handleMessage(msg, "https://api.telegram.org")
}
