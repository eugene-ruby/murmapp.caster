package telegramout_test

import (
	"context"
	"testing"

	"github.com/eugene-ruby/xconnect/redisstore"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"murmappcaster/internal/config"
	"murmappcaster/internal/telegramout"
	casterpb "murmappcaster/proto"
)

// ❌ proto.Unmarshal should fail
func Test_HandleMessageOut_invalid_proto(t *testing.T) {
	handler := &telegramout.OutboundHandler{
		Config: &config.Config{},
		Store:  redisstore.New(redisstore.NewMockClient()),
	}
	telegramout.HandleMessageOut([]byte("not-a-proto"), handler)
	// Expect: no panic
}

// ❌ Decryption of EncryptedApiKeyBot fails
func Test_HandleMessageOut_decryption_api_key_fail(t *testing.T) {
	msg := &casterpb.SendMessageRequest{
		EncryptedApiKeyBot: []byte("not-encrypted-data"),
		EncryptedPayload:   []byte("payload"),
		ApiEndpoint:        "send",
	}
	data, _ := proto.Marshal(msg)

	handler := &telegramout.OutboundHandler{
		Config: &config.Config{
			Encryption: config.EncryptionConfig{
				PayloadEncryptionKey: []byte("12345678901234567890123456789012"),
			},
		},
		Store: redisstore.New(redisstore.NewMockClient()),
	}
	telegramout.HandleMessageOut(data, handler)
	// Expect graceful fail
}

// ❌ Redis returns no value (XID not found)
func Test_HandleMessageOut_missing_xid_in_redis(t *testing.T) {
	payloadKey := []byte("12345678901234567890123456789012")
	secretBotKey := []byte("02345678901234567890123456789012")
	payload := `{"chat_id":"__XID:abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd__","text":"hello"}`

	encBotKey, _ := xsecrets.EncryptBytesWithKey([]byte("test-api-key"), secretBotKey)
	encPayload, _ := xsecrets.EncryptBytesWithKey([]byte(payload), payloadKey)

	msg := &casterpb.SendMessageRequest{
		EncryptedApiKeyBot: encBotKey,
		EncryptedPayload:   encPayload,
		ApiEndpoint:        "send",
	}
	data, _ := proto.Marshal(msg)

	var called bool
	telegramout.OverrideSendToTelegram(func(ctx context.Context, req *telegramout.OutgoingTelegramRequest) error {
		called = true
		return nil
	})
	defer telegramout.ResetSendToTelegram()

	handler := &telegramout.OutboundHandler{
		Config: &config.Config{
			Encryption: config.EncryptionConfig{
				SecretSalt:             []byte("salt123"),
				PayloadEncryptionKey:   payloadKey,
				SecretBotEncryptionKey: secretBotKey,
			},
		},
		Store: redisstore.New(redisstore.NewMockClient()),
	}

	telegramout.HandleMessageOut(data, handler)
	require.True(t, called, "Telegram should still be called even if XID is missing")
}
