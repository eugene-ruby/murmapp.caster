package registration

import (
	"fmt"
	"log"

	"github.com/eugene-ruby/xconnect/rabbitmq"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"google.golang.org/protobuf/proto"
	"murmapp.caster/internal/config"
	casterpb "murmapp.caster/proto"
)

type OutboundHandler struct {
	Config  *config.Config
	Channel rabbitmq.Channel
}

var HandleRegistration = func(body []byte, outboundHandler *OutboundHandler) {
	webhookHost := outboundHandler.Config.WebhookHost
	payloadEncryptionKey := outboundHandler.Config.Encryption.PayloadEncryptionKey
	telegramAPI := outboundHandler.Config.TelegramAPI
	secretSalt := outboundHandler.Config.Encryption.SecretSalt

	var req casterpb.RegisterWebhookRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		log.Printf("[registrations] ❌ failed to unmarshal protobuf: %v", err)
		return
	}

	log.Printf("[registrations] 📅 received registration request for botID: %s", req.BotId)

	secretToken := GenerateSecretToken()
	webhookID := ComputeWebhookID(secretToken, string(secretSalt))
	webhookURL := fmt.Sprintf("%s/api/webhook/%s", webhookHost, webhookID)

	// "payload"
	decryptApiKey, err := xsecrets.DecryptBytesWithKey(req.ApiKeyBot, payloadEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt bot api key: %v", err)
		return
	}

	if err := RegisterTelegramWebhook(string(decryptApiKey), webhookURL, secretToken, telegramAPI); err != nil {
		log.Printf("[registrations] ❌ webhook registration failed: %v", err)
		return
	}

	if err := registeredPush(req.BotId, webhookID, decryptApiKey, outboundHandler); err != nil {
		log.Printf("[caster] ❌ failed to push registered bot: %v", err)
		return
	}
}

func registeredPush(botID, webhookID string, decryptApiKey []byte, outboundHandler *OutboundHandler) error {
	channel := outboundHandler.Channel
	secretBotEncryptionKey := outboundHandler.Config.Encryption.SecretBotEncryptionKey

	encryptedApiKeyBot, err := xsecrets.EncryptBytesWithKey(decryptApiKey, secretBotEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ encryption failed: %v", err)
		return err
	}

	resp := &casterpb.RegisterWebhookResponse{
		BotId:              botID,
		EncryptedApiKeyBot: encryptedApiKeyBot,
		WebhookId:          webhookID,
	}

	body, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("[registrations] ❌ failed to marshal response: %v", err)
		return err
	}

	if err := channel.Publish("murmapp", "webhook.registered", body); err != nil {
		log.Printf("[caster] ❌ failed to publish to MQ: %v", err)
		return err
	}

	log.Printf("[registrations] ✅ registered webhookID: %s for botID: %s", webhookID, botID)
	return nil
}
