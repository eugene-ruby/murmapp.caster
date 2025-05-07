package telegramout

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/eugene-ruby/xconnect/redisstore"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"google.golang.org/protobuf/proto"
	"murmapp.caster/internal/config"
	casterpb "murmapp.caster/proto"
)

type OutboundHandler struct {
	Config *config.Config
	Store  *redisstore.Store
	DB     *sql.DB
}

// HandleEncryptedRequest handles a raw protobuf-encoded and encrypted SendMessageRequest.
func HandleMessageOut(body []byte, outboundHandler *OutboundHandler) {
	secretBotEncryptionKey := outboundHandler.Config.Encryption.SecretBotEncryptionKey
	payloadEncryptionKey := outboundHandler.Config.Encryption.PayloadEncryptionKey
	telegramIdEncryptionKey := outboundHandler.Config.Encryption.TelegramIdEncryptionKey
	telegramAPI := outboundHandler.Config.TelegramAPI

	var req casterpb.SendMessageRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		log.Printf("❌ Failed to decode proto: %v", err)
		return
	}

	apiKey, err := xsecrets.DecryptBytesWithKey(req.EncryptedApiKeyBot, secretBotEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt bot api key: %v", err)
		return
	}

	payload, err := xsecrets.DecryptBytesWithKey(req.EncryptedPayload, payloadEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt payload: %v", err)
		return
	}

	if !json.Valid(payload) {
		log.Printf("⚠️ Decrypted payload is not valid JSON!")
		return
	}

	outboundxID := &XIDPlaceholders{
		Redis:                   outboundHandler.Store,
		TelegramIdEncryptionKey: telegramIdEncryptionKey,
		TTL:                     10 * time.Minute,
	}

	payloadWithID, err := ReplaceXIDPlaceholders(payload, outboundxID)
	if err != nil {
		log.Printf("[caster] ❌ failed to replace telegram_id: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	requestAPI := &OutgoingTelegramRequest{
		ApiKey:      string(apiKey),
		Endpoint:    req.ApiEndpoint,
		TelegramAPI: telegramAPI,
		Payload:     payloadWithID,
	}

	if err := sendToTelegram(ctx, requestAPI); err != nil {
		log.Printf("📡 Telegram send failed: %v", err)
	}
}
