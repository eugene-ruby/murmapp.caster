package internal

import (
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

var HendlerRegistration = func(body []byte, mq Publisher) {
	var req casterpb.RegisterWebhookRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		log.Printf("[registrations] ❌ failed to unmarshal protobuf: %v", err)
		return
	}

	log.Printf("[registrations] 📅 received registration request for botID: %s", req.BotId)

	secretToken := GenerateSecretToken()
	webhookID := ComputeWebhookID(secretToken)
	webhookURL := fmt.Sprintf("%s/api/webhook/%s", WebhookHost, webhookID)

	decryptApiKey, err := DecryptWithKey(req.ApiKeyBot, PayloadEncryptionKey)
	if err != nil {
		log.Printf("[caster] ❌ failed to decrypt bot api key: %v", err)
		return
	}

	if err := RegisterTelegramWebhook(decryptApiKey, webhookURL, secretToken); err != nil {
		log.Printf("[registrations] ❌ webhook registration failed: %v", err)
		return
	}

	if err := registeredPush(req.BotId, webhookID, decryptApiKey, mq); err != nil {
		log.Printf("[caster] ❌ failed to push registered bot: %v", err)
		return
	}
}

func registeredPush(botID, webhookID, decryptApiKey string, mq Publisher) error {
	encryptedApiKeyBot, err := EncryptWithKey([]byte(decryptApiKey), SecretBotEncryptionKey)
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

	if err := mq.Publish("murmapp", "webhook.registered", body); err != nil {
		log.Printf("[caster] ❌ failed to publish to MQ: %v", err)
		return err
	}

	if err != nil {
		log.Printf("[registrations] ❌ publish error: %v", err)
		return err
	} else {
		log.Printf("[registrations] ✅ registered webhookID: %s for botID: %s", webhookID, botID)
	}

	return nil
}
