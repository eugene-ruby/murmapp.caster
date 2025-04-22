package internal

import (
	"testing"
	"sync"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func TestHendlerMsgOut_CallsHandler(t *testing.T) {
	var called bool
	var wg sync.WaitGroup
	wg.Add(1)

	// mock handler
	originalHandler := HendlerMessageOut
	defer func() { HendlerMessageOut = originalHandler }()

	HendlerMessageOut = func(body []byte, apiBase string) {
		defer wg.Done()
		called = true
		require.Equal(t, []byte("test-msg-out"), body)
		require.Equal(t, "https://api.example.com", apiBase)
	}

	// set global TelegramAPI
	TelegramAPI = "https://api.example.com"

	// simulate one delivery
	msgChan := make(chan amqp.Delivery, 1)
	msgChan <- amqp.Delivery{
		Body:       []byte("test-msg-out"),
		RoutingKey: "telegram.messages.out",
	}
	close(msgChan)

	// run handler
	HendlerMsgOut(msgChan, "test-queue")
	wg.Wait()

	require.True(t, called, "expected handler to be called")
}
