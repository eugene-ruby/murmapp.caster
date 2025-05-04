package telegramout

import (
	"fmt"
	"log"
	"net/http"
	"context"
	"strings"
)

type OutgoingTelegramRequest struct {
	ApiKey string
	Endpoint string
	TelegramAPI string
	Payload [] byte
}

// SendToTelegram posts a JSON payload to the specified Telegram API endpoint.
func SendToTelegram(ctx context.Context, request *OutgoingTelegramRequest) error {
	url := fmt.Sprintf("%s/bot%s/%s", request.TelegramAPI, request.ApiKey, request.Endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(request.Payload)))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram http error: %w", err)
	}
	defer resp.Body.Close()

	safeURL := fmt.Sprintf("%s/bot[redacted]/%s", request.TelegramAPI, request.Endpoint)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("❌ Telegram API %s returned status %d", safeURL, resp.StatusCode)
	} else {
		log.Printf("✅ Telegram API response: %s | → %s", resp.Status, safeURL)
	}

	return nil
}

var sendToTelegram = SendToTelegram

func OverrideSendToTelegram(fn func(context.Context, *OutgoingTelegramRequest) error) {
	sendToTelegram = fn
}

func ResetSendToTelegram() {
	sendToTelegram = SendToTelegram
}
