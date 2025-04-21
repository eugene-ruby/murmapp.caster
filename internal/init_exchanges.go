package internal

import (
	"github.com/streadway/amqp"
	"log"
)

// InitExchanges declares all topic exchanges used by the system
func InitExchanges(ch *amqp.Channel) error {
    // declare the exchange (just in case)
    exchange := "murmapp"

	err := ch.ExchangeDeclare(
			exchange,
			"topic", // exchange type
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			log.Printf("failed to declare exchange %s: %v", exchange, err)
			return err
		}
		log.Printf("exchange declared: %s", exchange)

	return nil
}
