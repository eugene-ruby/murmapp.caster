package internal

import (
    "bytes"
    "log"
    "net/http"

    "google.golang.org/protobuf/proto"
    "murmapp.caster/proto"
)

func handleMessage(body []byte, apiBase string) {
    var req casterpb.SendMessageRequest
    if err := proto.Unmarshal(body, &req); err != nil {
        log.Printf("failed to decode proto: %v", err)
        return
    }

    url := apiBase + "/" + req.ApiEndpoint
    httpReq, err := http.NewRequest("POST", url, bytes.NewReader(req.RawBody))
    if err != nil {
        log.Printf("failed to create request: %v", err)
        return
    }

    httpReq.Header.Set("Authorization", "Bearer "+req.BotApiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        log.Printf("failed to send to Telegram API: %v", err)
        return
    }
    defer resp.Body.Close()

    log.Printf("Telegram API response: %s %s", resp.Status, url)
}
