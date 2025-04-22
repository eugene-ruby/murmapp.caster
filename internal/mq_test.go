package internal_test

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"murmapp.caster/internal"
)

// MockCh implements the internal.Channel interface for testing
type MockCh struct {
	ExchangeDeclared bool
	Published        bool
	Closed           bool
}

func (m *MockCh) ExchangeDeclare(name, kind string, durable, autoDelete, internalFlag, noWait bool, args amqp.Table) error {
	m.ExchangeDeclared = true
	return nil
}

func (m *MockCh) Publish(exchange, routingKey string, body []byte) error {
	m.Published = true
	return nil
}

func (m *MockCh) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{Name: name}, nil
}

func (m *MockCh) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return nil
}

func (m *MockCh) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	ch := make(chan amqp.Delivery)
	close(ch)
	return ch, nil
}

func (m *MockCh) Close() error {
	m.Closed = true
	return nil
}

func TestMQPublisher_PublishAndClose(t *testing.T) {
	mock := &MockCh{}
	mq := &internal.MQPublisher{}
	mq.SetChannel(mock)

	err := mq.Publish("murmapp", "some.key", []byte("hello"))
	require.NoError(t, err)
	require.True(t, mock.Published)

	err = mq.Close()
	require.NoError(t, err)
	require.True(t, mock.Closed)
}

func TestWrapAMQPChannel_ImplementsAll(t *testing.T) {
	// Этот тест проверяет, что WrapAMQPChannel возвращает объект,
	// который реализует интерфейс internal.Channel без ошибок (компиляция = успех)
	var ch internal.Channel = internal.WrapAMQPChannel(&amqp.Channel{})
	require.NotNil(t, ch)
}
