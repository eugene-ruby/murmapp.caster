package internal

import (
    "log"

    "github.com/streadway/amqp"
)

func StartConsumerMsgOUT(ch *amqp.Channel, telegramAPI string) error {

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

    go func() {
    for d := range msgs {
        log.Printf(
            "📩 Message received | queue: %s | routing_key: %s | size: %d bytes",
            q.Name, d.RoutingKey, len(d.Body),
        )
        go hendlerMessageOut(d.Body, telegramAPI)
    }
    }()

    log.Println("🗣️ caster is running...")
    select {}
}
