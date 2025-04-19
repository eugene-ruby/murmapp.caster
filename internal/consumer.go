package internal

import (
    "log"

    "github.com/streadway/amqp"
)

func StartConsumer(rabbitURL, telegramAPI string) error {
    conn, err := amqp.Dial(rabbitURL)
    if err != nil {
        return err
    }
    ch, err := conn.Channel()
    if err != nil {
        return err
    }

    // declare the exchange (just in case)
    err = ch.ExchangeDeclare(
        "murmapp", // name
        "topic",                // type
        true,                   // durable
        false,                  // auto-deleted
        false,                  // internal
        false,                  // no-wait
        nil,                    // args
    )
    if err != nil {
        return err
    }

    // declare the queue
    q, err := ch.QueueDeclare(
        "murmapp.caster.telegram.messages.out", true, false, false, false, nil,
    )
    if err != nil {
        return err
    }

    // bind by routing_key
    err = ch.QueueBind(
        q.Name, "telegram.messages.out", "murmapp", false, nil,
    )
    if err != nil {
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
        go handleMessage(d.Body, telegramAPI)
    }
    }()

    log.Println("🗣️ caster is running...")
    select {} // block forever
}
