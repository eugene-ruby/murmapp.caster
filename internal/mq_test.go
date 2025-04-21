package internal_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"murmapp.caster/internal"
)

// MockAMQPChannel mocks the amqp.Channel
type MockAMQPChannel struct {
	PublishedExchange   string
	PublishedRoutingKey string
	PublishedBody       []byte
}

func (m *MockAMQPChannel) Publish(exchange, routingKey string, mandatory, immediate bool, msg interface{}) error {
	publishing := msg.(struct {
		ContentType string
		Body        []byte
	})
	m.PublishedExchange = exchange
	m.PublishedRoutingKey = routingKey
	m.PublishedBody = publishing.Body
	return nil
}

// MockPublisher implements internal.Publisher using the mock channel
type MockPublisher struct {
	Channel *MockAMQPChannel
}

func (m *MockPublisher) Publish(exchange, routingKey string, body []byte) error {
	return m.Channel.Publish(exchange, routingKey, false, false, struct {
		ContentType string
		Body        []byte
	}{
		ContentType: "application/octet-stream",
		Body:        body,
	})
}

func TestMockPublisher_Publish(t *testing.T) {
	mockChannel := &MockAMQPChannel{}
	publisher := &MockPublisher{Channel: mockChannel}

	err := publisher.Publish("murmapp", "test.key", []byte("hello"))
	require.NoError(t, err)

	require.Equal(t, "murmapp", mockChannel.PublishedExchange)
	require.Equal(t, "test.key", mockChannel.PublishedRoutingKey)
	require.Equal(t, []byte("hello"), mockChannel.PublishedBody)
}

func TestInitMQ_Success(t *testing.T) {
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672")

	mq, err := internal.InitMQ()
	require.NoError(t, err)
	require.NotNil(t, mq)
	defer mq.Close()

	ch := mq.GetChannel()
	require.NotNil(t, ch)
}

func TestPublish_Success(t *testing.T) {
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672")

	mq, err := internal.InitMQ()
	require.NoError(t, err)
	defer mq.Close()

	err = mq.Publish("amq.direct", "test.publish", []byte("test message"))
	require.NoError(t, err)
}
