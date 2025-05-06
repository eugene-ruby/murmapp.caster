package registration_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eugene-ruby/xconnect/rabbitmq/mocks"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"murmapp.caster/internal/config"
	"murmapp.caster/internal/registration"
	casterpb "murmapp.caster/proto"
)

func Test_HandleRegistration_success(t *testing.T) {
	// Setup a fake Telegram API server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)
		require.Contains(t, payload, "url")
		require.Contains(t, payload, "secret_token")

		w.WriteHeader(http.StatusOK)
	}))
	defer fakeServer.Close()

	// Setup a mock channel
	mockChannel := mocks.NewMockChannel()

	payloadEncryptionKey := []byte("32323232323232323232323232323232")
	secretBotEncryptionKey := []byte("02323232323232323232323232323232")

	// 🏗 Build handler
	handler := &registration.OutboundHandler{
		Config: &config.Config{
			TelegramAPI: fakeServer.URL,
			WebhookHost: "https://example.com/webhook",
			Encryption: config.EncryptionConfig{
				SecretSalt: []byte("test_salt"),
				PayloadEncryptionKey:   payloadEncryptionKey,
				SecretBotEncryptionKey: secretBotEncryptionKey,
			},
		},
		Channel: mockChannel,
	}

	// Prepare a test RegisterWebhookRequest
	botID := "test_bot"
	apiKey := "real_bot_api_key"

	encryptedApiKey, err := xsecrets.EncryptBytesWithKey([]byte(apiKey), payloadEncryptionKey)
	require.NoError(t, err)

	req := &casterpb.RegisterWebhookRequest{
		BotId:     botID,
		ApiKeyBot: encryptedApiKey,
	}
	body, err := proto.Marshal(req)
	require.NoError(t, err)

	// Call the function
	registration.HandleRegistration(body, handler)

	// Assertions
	require.Len(t, mockChannel.PublishedMessages, 1, "should publish one message")

	published := mockChannel.PublishedMessages[0].Body

	// Unmarshal and check published RegisterWebhookResponse
	var resp casterpb.RegisterWebhookResponse
	err = proto.Unmarshal(published, &resp)
	require.NoError(t, err)

	require.Equal(t, botID, resp.BotId)
	require.NotEmpty(t, resp.EncryptedApiKeyBot)
	require.NotEmpty(t, resp.WebhookId)

	original_bot_key, _ := xsecrets.DecryptBytesWithKey(resp.EncryptedApiKeyBot, handler.Config.Encryption.SecretBotEncryptionKey)
	require.Equal(t, original_bot_key, []byte(apiKey))
}

func Test_HandleRegistration_invalid_protobuf(t *testing.T) {
	mockChannel := mocks.NewMockChannel()

	// Pass invalid protobuf data
	invalidBody := []byte("this is not valid protobuf")

	// Set the minimum config
	handler := &registration.OutboundHandler{
    Config: &config.Config{
        WebhookHost: "https://example.com",
        Encryption: config.EncryptionConfig{
            PayloadEncryptionKey: []byte("12345678901234567890123456789012"),
        },
    },
    Channel: mocks.NewMockChannel(),
	}

	registration.HandleRegistration(invalidBody, handler)

	// Should not publish anything
	require.Len(t, mockChannel.PublishedMessages, 0, "should not publish if protobuf unmarshal failed")
}

func Test_HandleRegistration_decrypt_failure(t *testing.T) {
	mockChannel := mocks.NewMockChannel()

	// Создаем нормальный protobuf, но неправильный зашифрованный ключ
	req := &casterpb.RegisterWebhookRequest{
		BotId:     "test_bot",
		ApiKeyBot: []byte("invalid-encrypted-data"),
	}
	body, err := proto.Marshal(req)
	require.NoError(t, err)

	// Setup config with wrong PayloadEncryptionKey
	handler := &registration.OutboundHandler{
		Config: &config.Config{
			WebhookHost: "https://example.com",
			Encryption: config.EncryptionConfig{
				PayloadEncryptionKey: []byte("wrong-key-32-bytes-wrong-key-32---"),
			}},
		Channel: mockChannel,
	}

	registration.HandleRegistration(body, handler)

	// Should not publish anything
	require.Len(t, mockChannel.PublishedMessages, 0, "should not publish if decryption failed")
}

func Test_HandleRegistration_register_webhook_failure(t *testing.T) {
	mockChannel := mocks.NewMockChannel()

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer fakeServer.Close()

	payloadKey := []byte("32323232323232323232323232323232")
	encryptedApiKey, err := xsecrets.EncryptBytesWithKey([]byte("fakeapikey"), payloadKey)
	require.NoError(t, err)

	req := &casterpb.RegisterWebhookRequest{
		BotId:     "test_bot",
		ApiKeyBot: encryptedApiKey,
	}
	body, err := proto.Marshal(req)
	require.NoError(t, err)

	handler := &registration.OutboundHandler{
		Channel: mockChannel,
		Config: &config.Config{
			WebhookHost: "https://example.com",
			TelegramAPI: fakeServer.URL,
			Encryption: config.EncryptionConfig{
				PayloadEncryptionKey:   []byte("32323232323232323232323232323232"),
				SecretBotEncryptionKey: []byte("32323232323232323232323232323232"),
			},
		},
	}

	registration.HandleRegistration(body, handler)

	// Should not publish anything
	require.Len(t, mockChannel.PublishedMessages, 0, "should not publish if webhook registration failed")
}

func Test_HandleRegistration_publish_failure(t *testing.T) {
	mockChannel := mocks.NewMockChannel()
	mockChannel.PublishErr = errors.New("publish fail")

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer fakeServer.Close()

	payloadKey := []byte("32323232323232323232323232323232")
	encryptedApiKey, err := xsecrets.EncryptBytesWithKey([]byte("fakeapikey"), payloadKey)
	require.NoError(t, err)

	req := &casterpb.RegisterWebhookRequest{
		BotId:     "test_bot",
		ApiKeyBot: encryptedApiKey,
	}
	body, err := proto.Marshal(req)
	require.NoError(t, err)

	handler := &registration.OutboundHandler{
		Channel: mockChannel,
		Config: &config.Config{
			WebhookHost: "https://example.com",
			TelegramAPI: fakeServer.URL,
			Encryption: config.EncryptionConfig{
				PayloadEncryptionKey:   []byte("32323232323232323232323232323232"),
				SecretBotEncryptionKey: []byte("32323232323232323232323232323232"),
			},
		},
	}

	// Should not panic or crash even if Publish fails
	require.NotPanics(t, func() {
		registration.HandleRegistration(body, handler)
	})
}
