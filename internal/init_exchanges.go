package internal

import (
	"log"
)

// InitExchanges declares all topic exchanges used by the system
func InitExchanges(mq *MQPublisher) error {
    // declare the exchange (just in case)
	ch := mq.ch
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
