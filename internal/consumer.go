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

    // Объявляем exchange (на всякий случай)
    err = ch.ExchangeDeclare(
        "murmapp.messages.out", // name
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

    // Объявляем очередь
    q, err := ch.QueueDeclare(
        "caster_telegram_send", true, false, false, false, nil,
    )
    if err != nil {
        return err
    }

    // Биндим по routing_key
    err = ch.QueueBind(
        q.Name, "telegram.send", "murmapp.messages.out", false, nil,
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
            go handleMessage(d.Body, telegramAPI)
        }
    }()

    log.Println("caster is running...")
    select {} // block forever
}
