package telegramout_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"fmt"

	"github.com/stretchr/testify/require"
	"murmapp.caster/internal/telegramout"
)

func Test_SendToTelegram_success(t *testing.T) {
	var called bool
	var receivedBody string
	var receivedContentType string
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		receivedContentType = r.Header.Get("Content-Type")
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := &telegramout.OutgoingTelegramRequest{
		ApiKey:      "FAKE_TOKEN",
		Endpoint:    "sendMessage",
		TelegramAPI: server.URL,
		Payload:     []byte(`{"chat_id":123,"text":"hello"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := telegramout.SendToTelegram(ctx, req)
	require.NoError(t, err)

	require.True(t, called, "server should be called")
	require.Equal(t, `{"chat_id":123,"text":"hello"}`, receivedBody)
	require.Equal(t, "application/json", receivedContentType)
	require.Equal(t, "POST", receivedMethod)
}

func Test_SendToTelegram_failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}))
	defer server.Close()

	req := &telegramout.OutgoingTelegramRequest{
		ApiKey:      "FAKE_TOKEN",
		Endpoint:    "sendMessage",
		TelegramAPI: server.URL,
		Payload:     []byte(`{"chat_id":123,"text":"hello"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := telegramout.SendToTelegram(ctx, req)
	require.Error(t, err)
	expectedURL := fmt.Sprintf("%s/bot[redacted]/sendMessage", server.URL)
	require.Contains(t, err.Error(), expectedURL)
}
