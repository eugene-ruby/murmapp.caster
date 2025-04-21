package internal

import (
	"fmt"
	"os"
	"time"
    "log"

	"github.com/streadway/amqp"
)

var RabbitURL string

type MQPublisher struct {
	conn    *amqp.Connection
	ch Channel
	channel *amqp.Channel
}

type Publisher interface {
	Publish(exchange, routingKey string, body []byte) error
}

func (mq *MQPublisher) GetChannel() *amqp.Channel {
	return mq.channel
}

func (mq *MQPublisher) SetChannel(ch Channel) {
	mq.ch = ch
}

type Channel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	Publish(exchange, routingKey string, body []byte) error

	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error

	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)

	Close() error
}


func InitMQ() (*MQPublisher, error) {
	var conn *amqp.Connection
	var err error

	RabbitURL := os.Getenv("RABBITMQ_URL")
	if RabbitURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL env var not set")
	}

	for retries := 0; retries < 30; retries++ {
		conn, err = amqp.Dial(RabbitURL)
		if err == nil {
			log.Printf("✅ Connected to RabbitMQ on attempt %d", retries+1)
			break
		}
		log.Printf("❌ Failed to connect to RabbitMQ (attempt %d/30): %v", retries+1, err)
		time.Sleep(5 * time.Second)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &MQPublisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *MQPublisher) Publish(exchange, routingKey string, body []byte) error {
	return p.channel.Publish(
		exchange,
		routingKey,
		false, false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        body,
		})
}

func (p *MQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
