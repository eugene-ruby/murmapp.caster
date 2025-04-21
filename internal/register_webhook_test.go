package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterTelegramWebhook(t *testing.T) {
	// mock Telegram server
	var capturedBody map[string]interface{}
	var capturedPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(bodyBytes, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// override global TelegramAPI value
	TelegramAPI = ts.URL

	apiKey := "123456:ABC-DEF"
	webhookURL := "https://example.com/api/webhook/abc123"
	secretToken := "supersecrettoken"

	err := RegisterTelegramWebhook(apiKey, webhookURL, secretToken)
	require.NoError(t, err)

	// validate request path
	require.Equal(t, "/bot123456:ABC-DEF/setWebhook", capturedPath)

	// validate body
	require.Equal(t, webhookURL, capturedBody["url"])
	require.Equal(t, secretToken, capturedBody["secret_token"])
}
