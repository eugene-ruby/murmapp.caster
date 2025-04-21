package internal

import (
	"log"

	"github.com/streadway/amqp"
)

func StartRegistrationConsumer(ch *amqp.Channel) error {
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

    go func() {
    for d := range msgs {
        log.Printf(
            "📩 Message received | queue: %s | routing_key: %s | size: %d bytes",
            q.Name, d.RoutingKey, len(d.Body),
        )
        go hendlerRegistration(d.Body, ch)
    }
    }()

	log.Println("[caster.registrations] 📖 consumer started and listening...")
    select {}
}
