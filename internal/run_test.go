package internal_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/streadway/amqp"
	"murmapp.caster/internal"
)

type MockChannel struct {
	CalledExchange bool
}

func (m *MockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	m.CalledExchange = true
	return nil
}

func (m *MockChannel) Close() error  { return nil }
func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{Name: name}, nil
}
func (m *MockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return nil
}
func (m *MockChannel) Consume(name, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	ch := make(chan amqp.Delivery)
	close(ch)
	return ch, nil
}
func (m *MockChannel) Publish(exchange, routingKey string, body []byte) error {
	return nil
}

func TestInitExchangesFunc_WrapsInitExchanges(t *testing.T) {
	mq := &internal.MQPublisher{}
	mq.SetChannel(&MockChannel{})

	err := internal.InitExchanges(mq)
	require.NoError(t, err)
	// require.True(t, mockCh.ExchangeDeclared)
}

func TestStartRegistrationConsumerFunc_DoesNotPanic(t *testing.T) {
	mq := &internal.MQPublisher{}
	mq.SetChannel(&MockChannel{})

	go func() {
		_ = internal.StartRegistrationConsumer(mq)
	}()
}

func TestStartConsumerMsgOutFunc_DoesNotPanic(t *testing.T) {
	mq := &internal.MQPublisher{}
	mq.SetChannel(&MockChannel{})	

	go func() {
		_ = internal.StartConsumerMsgOut(mq, "https://api.telegram.org")
	}()
}

func TestRun_HTTPBoots(t *testing.T) {
	// mock InitMQ and consumers

	internal.InitMQFunc = func() (*internal.MQPublisher, error) {
		return &internal.MQPublisher{}, nil
	}

	internal.InitExchangesFunc = func(mq *internal.MQPublisher) error {
		return nil
	}	

	internal.StartRegistrationConsumerFunc = func(mq *internal.MQPublisher) error {
		return nil
	}

	internal.StartConsumerMsgOutFunc = func(mq *internal.MQPublisher, api string) error {
		return nil
	}

	// Minimal env vars for InitEncryptionKey
	_ = os.Setenv("APP_PORT", "3999")
	_ = os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	_ = os.Setenv("BOT_ENCRYPTION_KEY", "12345678901234567890123456789069")
	_ = os.Setenv("TELEGRAM_ID_ENCRYPTION_KEY", "12345678901234567890123456789012")
	_ = os.Setenv("WEB_HOOK_HOST", "https://example.com")
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@blalba.io:5672")

	// Run in goroutine so we can test healthz
	go func() {
		_ = internal.Run()
	}()

	time.Sleep(1 * time.Second) // Give server time to start

	res, err := http.Get("http://localhost:3999/healthz")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	originalInitMQ := internal.InitMQFunc
	defer func() { internal.InitMQFunc = originalInitMQ }()
}
