package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

type MockChannel struct {
	Called             bool
	LastExchange       string
	LastRoutingKey     string
	LastBody           []byte
	PublishShouldError bool
}

func (m *MockChannel) Publish(exchange, routingKey string, body []byte) error {
	m.Called = true
	m.LastExchange = exchange
	m.LastRoutingKey = routingKey
	m.LastBody = body
	if m.PublishShouldError {
		return fmt.Errorf("mock publish failed")
	}
	return nil
}

// Stub methods
func (m *MockChannel) ExchangeDeclare(string, string, bool, bool, bool, bool, amqp.Table) error {
	return nil
}
func (m *MockChannel) QueueDeclare(string, bool, bool, bool, bool, amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{}, nil
}
func (m *MockChannel) QueueBind(string, string, string, bool, amqp.Table) error {
	return nil
}
func (m *MockChannel) Consume(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error) {
	return nil, nil
}
func (m *MockChannel) Close() error {
	return nil
}

func TestHandlerRegistration_ValidPayload(t *testing.T) {
	// Set up required encryption keys as environment variables
	_ = os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = os.Setenv("TELEGRAM_ID_ENCRYPTION_KEY", "12345678901234567890123456789012")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "12345678901234567890123456789069")
	_ = os.Setenv("TELEGRAM_API_URL", "https://api.example.com")
	_ = os.Setenv("WEB_HOOK_HOST", "https://myaip.example.com/api/webhook")
	_ = os.Setenv("SECRET_SALT", "blalbal321")

	// Initialize encryption keys from env
	if err := InitEncryptionKey(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := InitEnv(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Expected original values
	botID := "bot123"
	originalBotApiKey := "123456:ABC-DEF"

	// Capture request URL and body sent to the mock server
	var capturedURL string
	var capturedBody []byte

	// Create a mock Telegram API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		body, _ := io.ReadAll(r.Body)
		capturedBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	TelegramAPI = ts.URL

	// Encrypt both the API key and payload using the appropriate keys
	encApiKey, err := EncryptWithKey([]byte(originalBotApiKey), PayloadEncryptionKey)
	require.NoError(t, err)

	// Construct the protobuf message as would be published to RabbitMQ
	req := &casterpb.RegisterWebhookRequest{
		ApiKeyBot: encApiKey,
		BotId:     botID,
	}
	data, err := proto.Marshal(req)
	require.NoError(t, err)

	mockCh := &MockChannel{}
	mq := &MQPublisher{}
	mq.SetChannel(mockCh)

	HandlerRegistration(data, mq)

	// Validate that the resulting API request matches expected URL and body
	require.Equal(t, "/bot123456:ABC-DEF/setWebhook", capturedURL)
	// ✅ Ensures the request was sent to the correct Telegram endpoint using decrypted API key

	require.NoError(t, err)
	require.True(t, mockCh.Called)
	require.Equal(t, "murmapp", mockCh.LastExchange)
	require.Equal(t, "webhook.registered", mockCh.LastRoutingKey)
	// ✅ Confirms that a message was published to the correct exchange and routing key

	var resp casterpb.RegisterWebhookResponse
	err = proto.Unmarshal(mockCh.LastBody, &resp)
	require.NoError(t, err)
	require.Equal(t, botID, resp.BotId)
	// ✅ Validates that the pushed protobuf contains the correct bot ID

	decryptSecretApiKey, err := DecryptWithKey([]byte(resp.EncryptedApiKeyBot), SecretBotEncryptionKey)
	require.NoError(t, err)
	require.Equal(t, originalBotApiKey, decryptSecretApiKey)
	// ✅ Ensures that the EncryptedApiKeyBot in the push can be decrypted and matches original

	var body map[string]interface{}
	err = json.Unmarshal(capturedBody, &body)
	require.NoError(t, err)
	urlStr, ok := body["url"].(string)
	require.True(t, ok, "expected 'url' field in JSON body")
	parts := strings.Split(urlStr, "/")
	webhookID := parts[len(parts)-1]
	require.Equal(t, resp.WebhookId, webhookID)
	// ✅ Parses the webhookID from the sent request and matches it with the protobuf payload
}
