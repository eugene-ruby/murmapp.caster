package telegramout_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/eugene-ruby/xconnect/redisstore"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"murmapp.caster/internal/config"
	"murmapp.caster/internal/telegramout"
	casterpb "murmapp.caster/proto"
)

func Test_HandleMessageOut_success_with_XID(t *testing.T) {
	ctx := context.Background()

	// 🔐 Generate RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	// 📦 Telegram ID to encrypt
	telegramID := []byte("123456789")

	// 🔒 Encrypt telegram ID
	encryptedBytes, err := xsecrets.RSAEncryptBytes(pubKey, telegramID)
	require.NoError(t, err)

	// 📦 Marshal proto.TelegramIDStore
	storeProto := &casterpb.TelegramIdStore{
		Version:          "v1",
		EncryptedPayload: encryptedBytes,
	}
	rawStore, err := proto.Marshal(storeProto)
	require.NoError(t, err)

	// 🧂 Hash telegram ID with salt
	salt := "salt123"
	h := sha256.New()
	h.Write(telegramID)
	h.Write([]byte(salt))
	hash := h.Sum(nil)
	hashHex := fmt.Sprintf("%x", hash)

	// 🧠 Mock Redis
	mock := redisstore.NewMockClient()
	store := redisstore.New(mock)
	err = store.Set(ctx, hashHex, string(rawStore), time.Minute)
	require.NoError(t, err)

	// 📄 Payload with __XID:{hash}
	payload := `{"chat_id":"__XID:` + hashHex + `__","text":"hello"}`
	payloadKey := []byte("12345678901234567890123456789012")
	secretBotKey := []byte("02345678901234567890123456789012")

	encAPI, err := xsecrets.EncryptBytesWithKey([]byte("test-api-key"), secretBotKey)
	require.NoError(t, err)
	encPayload, err := xsecrets.EncryptBytesWithKey([]byte(payload), payloadKey)
	require.NoError(t, err)

	// 📦 Create SendMessageRequest proto
	msg := &casterpb.SendMessageRequest{
		EncryptedApiKeyBot: encAPI,
		EncryptedPayload:   encPayload,
		ApiEndpoint:        "sendMessage",
	}
	data, err := proto.Marshal(msg)
	require.NoError(t, err)

	// 🧪 Capture outgoing Telegram request
	var capturedReq *telegramout.OutgoingTelegramRequest
	telegramout.OverrideSendToTelegram(func(_ context.Context, req *telegramout.OutgoingTelegramRequest) error {
		capturedReq = req
		return nil
	})
	defer telegramout.ResetSendToTelegram()

	// 🏗 Build handler
	handler := &telegramout.OutboundHandler{
		Config: &config.Config{
			TelegramAPI: "https://api.telegram.org/bot",
			Encryption: config.EncryptionConfig{
				SecretSalt: []byte(salt),
				PayloadEncryptionKey:    payloadKey,
				PrivateRSAEncryptionKey: privKey,
				SecretBotEncryptionKey: secretBotKey,
			},
		},
		Store: store,
	}

	// 🚀 Call handler
	telegramout.HandleMessageOut(data, handler)

	// ✅ Assertions
	require.NotNil(t, capturedReq)
	require.Equal(t, "test-api-key", capturedReq.ApiKey)
	require.Equal(t, "https://api.telegram.org/bot", capturedReq.TelegramAPI)
	require.Equal(t, "sendMessage", capturedReq.Endpoint)

	var parsed map[string]any
	err = json.Unmarshal(capturedReq.Payload, &parsed)
	require.NoError(t, err)
	require.Equal(t, "123456789", parsed["chat_id"])
	require.Equal(t, "hello", parsed["text"])
}
