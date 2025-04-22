package internal

import (
    "log"

    "github.com/streadway/amqp"
)

func StartConsumerMsgOut(mq *MQPublisher, telegramAPI string) error {
    ch := mq.ch
    q, err := ch.QueueDeclare("murmapp.caster.telegram.messages.out", true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(q.Name, "telegram.messages.out", "murmapp", false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

    go HendlerMsgOut(msgs, q.Name)

    log.Println("🗣️ caster is running...")
    select {}
}

func HendlerMsgOut(deliveries <-chan amqp.Delivery, queueName string) {
	for d := range deliveries {
		log.Printf("📩 Message received | queue: %s | routing_key: %s | size: %d bytes", queueName, d.RoutingKey, len(d.Body))
		go HendlerMessageOut(d.Body, TelegramAPI)
	}
}
