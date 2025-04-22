package internal

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Channel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	Publish(exchange, routingKey string, body []byte) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type MQPublisher struct {
	conn *amqp.Connection
	ch   Channel
}

func (p *MQPublisher) GetChannel() Channel {
	return p.ch
}

func (p *MQPublisher) SetChannel(ch Channel) {
	p.ch = ch
}

func (p *MQPublisher) Publish(exchange, routingKey string, body []byte) error {
	return p.ch.Publish(exchange, routingKey, body)
}

func (p *MQPublisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func InitMQ() (*MQPublisher, error) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL env var not set")
	}

	var conn *amqp.Connection
	var err error

	for retries := 0; retries < 30; retries++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			log.Printf("✅ Connected to RabbitMQ on attempt %d", retries+1)
			break
		}
		log.Printf("❌ Failed to connect to RabbitMQ (attempt %d/30): %v", retries+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to RabbitMQ after retries: %w", err)
	}

	amqpChannel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &MQPublisher{
		conn: conn,
		ch:   WrapAMQPChannel(amqpChannel),
	}, nil
}

// ===== Wrapper for *amqp.Channel to implement our Channel interface =====

type amqpChannelWrapper struct {
	raw *amqp.Channel
}

func WrapAMQPChannel(ch *amqp.Channel) Channel {
	return &amqpChannelWrapper{raw: ch}
}

func (a *amqpChannelWrapper) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return a.raw.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (a *amqpChannelWrapper) Publish(exchange, routingKey string, body []byte) error {
	return a.raw.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/octet-stream",
		Body:        body,
	})
}

func (a *amqpChannelWrapper) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return a.raw.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (a *amqpChannelWrapper) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return a.raw.QueueBind(name, key, exchange, noWait, args)
}

func (a *amqpChannelWrapper) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return a.raw.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (a *amqpChannelWrapper) Close() error {
	return a.raw.Close()
}
