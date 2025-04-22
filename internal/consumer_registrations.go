package internal

import (
	"log"

	"github.com/streadway/amqp"
)

func StartRegistrationConsumer(mq *MQPublisher) error {
	ch := mq.ch
	q, err := ch.QueueDeclare("murmapp.caster.webhook.registration", true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(q.Name, "webhook.registration", "murmapp", false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

    go HandleRegistrationMessages(msgs, mq, q.Name)

	log.Println("[caster.registrations] 📖 consumer started and listening...")
    select {}
}

func HandleRegistrationMessages(deliveries <-chan amqp.Delivery, mq *MQPublisher, queueName string) {
	for d := range deliveries {
		log.Printf("📩 Message received | queue: %s | routing_key: %s | size: %d bytes", queueName, d.RoutingKey, len(d.Body))
		go HendlerRegistration(d.Body, mq)
	}
}