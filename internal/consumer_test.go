package internal

import (
    "testing"
    "google.golang.org/protobuf/proto"
    "murmapp.caster/proto"
)

func TestUnmarshalSendMessageRequest(t *testing.T) {
    expected := &casterpb.SendMessageRequest{
        BotApiKey:   "123:abc",
        ApiEndpoint: "sendMessage",
        RawBody:     []byte(`{"chat_id":12345,"text":"Hello"}`),
    }

    bytesData, err := proto.Marshal(expected)
    if err != nil {
        t.Fatalf("failed to marshal proto: %v", err)
    }

    var actual casterpb.SendMessageRequest
    err = proto.Unmarshal(bytesData, &actual)
    if err != nil {
        t.Fatalf("failed to unmarshal proto: %v", err)
    }

    if actual.BotApiKey != expected.BotApiKey {
        t.Errorf("bot_api_key mismatch: got %s, want %s", actual.BotApiKey, expected.BotApiKey)
    }
    if actual.ApiEndpoint != expected.ApiEndpoint {
        t.Errorf("api_endpoint mismatch: got %s, want %s", actual.ApiEndpoint, expected.ApiEndpoint)
    }
    if string(actual.RawBody) != string(expected.RawBody) {
        t.Errorf("raw_body mismatch: got %s, want %s", actual.RawBody, expected.RawBody)
    }
}
