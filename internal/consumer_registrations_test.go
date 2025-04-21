package internal

import (
	"testing"
	"sync"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func TestHandleRegistrationMessages_CallsHandler(t *testing.T) {
	called := false

	mockMQ := &MockPublisher{}

	mockMsg := amqp.Delivery{
		Body:       []byte("test-body"),
		RoutingKey: "webhook.registration",
	}

	msgChan := make(chan amqp.Delivery, 1)
	msgChan <- mockMsg
	close(msgChan)

	originalHandler := HendlerRegistration
	defer func() { HendlerRegistration = originalHandler }()

	var wg sync.WaitGroup
	wg.Add(1)

	HendlerRegistration = func(body []byte, mq Publisher) {
		defer wg.Done()
		called = true
		require.Equal(t, []byte("test-body"), body)
		require.Equal(t, mockMQ, mq)
	}

	HandleRegistrationMessages(msgChan, mockMQ, "test-queue")
	// Wait until handler completes
	wg.Wait()
	require.True(t, called, "expected handler to be called")
}
