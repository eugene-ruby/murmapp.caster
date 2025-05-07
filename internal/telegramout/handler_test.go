package telegramout_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
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

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T, pgDNS string) *sql.DB {
	db, err := sql.Open("pgx", pgDNS)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`DELETE FROM telegram_id_map`)
	require.NoError(t, err)

	return db
}

func setupTestHandler(t *testing.T) *telegramout.OutboundHandler {
	cfg, _ := config.LoadConfig()

	// Mock Redis
	mock := redisstore.NewMockClient()
	store := redisstore.New(mock)

	db := setupTestDB(t, cfg.PostgreSQL.DSN)
	handler := &telegramout.OutboundHandler{
		DB:     db,
		Config: cfg,
		Store:  store,
	}

	return handler
}

func Test_HandleMessageOut_success_with_XID(t *testing.T) {
	handler := setupTestHandler(t)
	ctx := context.Background()

	// 📦 Telegram ID to encrypt
	telegramID := []byte("123456789")

	// 🔒 Encrypt telegram ID
	encryptedBytes, err := xsecrets.EncryptBytesWithKey(telegramID, handler.Config.Encryption.TelegramIdEncryptionKey)
	require.NoError(t, err)

	storeProto := &casterpb.TelegramIdStore{
		Version:          "v1",
		EncryptedPayload: encryptedBytes,
	}
	rawStore, err := proto.Marshal(storeProto)
	require.NoError(t, err)

	// 🧂 Hash telegram ID with salt
	salt := handler.Config.Encryption.SecretSalt
	h := sha256.New()
	h.Write(telegramID)
	h.Write(salt)
	hash := h.Sum(nil)
	hashHex := fmt.Sprintf("%x", hash)

	err = handler.Store.Set(ctx, hashHex, string(rawStore), time.Minute)
	require.NoError(t, err)

	// 📄 Payload with __XID:{hash}
	payload := `{"chat_id":"__XID:` + hashHex + `__","text":"hello"}`

	encAPI, err := xsecrets.EncryptBytesWithKey([]byte("test-api-key"), handler.Config.Encryption.SecretBotEncryptionKey)
	require.NoError(t, err)
	encPayload, err := xsecrets.EncryptBytesWithKey([]byte(payload), handler.Config.Encryption.PayloadEncryptionKey)
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

	// 🚀 Call handler
	telegramout.HandleMessageOut(data, handler)

	// ✅ Assertions
	require.NotNil(t, capturedReq)
	require.Equal(t, "test-api-key", capturedReq.ApiKey)
	require.Equal(t, "https://api.example.com", capturedReq.TelegramAPI)
	require.Equal(t, "sendMessage", capturedReq.Endpoint)

	var parsed map[string]any
	err = json.Unmarshal(capturedReq.Payload, &parsed)
	require.NoError(t, err)
	require.Equal(t, string(telegramID), parsed["chat_id"])
	require.Equal(t, "hello", parsed["text"])
}
