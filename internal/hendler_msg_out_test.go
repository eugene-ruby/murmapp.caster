package internal

import (
	"testing"
	"os"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

func TestHendlerMessageOut_ValidPayload(t *testing.T) {
	// Set up required encryption keys as environment variables
	_ = os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = os.Setenv("TELEGRAM_ID_ENCRYPTION_KEY", "12345678901234567890123456789012")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "12345678901234567890123456789069")

	// Initialize encryption keys from env
	err := InitEncryptionKey()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Expected original values
	originalJSON := `{"chat_id":123456,"text":"hi"}`
	originalApiKey := "123456:ABC-DEF"

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

	// Encrypt both the API key and payload using the appropriate keys
	encApiKey, err := EncryptWithKey([]byte(originalApiKey), SecretBotEncryptionKey)
	require.NoError(t, err)

	encPayload, err := EncryptWithKey([]byte(originalJSON), PayloadEncryptionKey)
	require.NoError(t, err)

	// Construct the protobuf message as would be published to RabbitMQ
	req := &casterpb.SendMessageRequest{
		EncryptedApiKeyBot: encApiKey,
		ApiEndpoint:        "sendMessage",
		EncryptedPayload:   encPayload,
	}
	data, err := proto.Marshal(req)
	require.NoError(t, err)

	// Execute the message handler which simulates processing a message from the queue
	HendlerMessageOut(data, ts.URL)

	// Validate that the resulting API request matches expected URL and body
	require.Equal(t, "/bot123456:ABC-DEF/sendMessage", capturedURL)
	require.JSONEq(t, originalJSON, string(capturedBody))
}
