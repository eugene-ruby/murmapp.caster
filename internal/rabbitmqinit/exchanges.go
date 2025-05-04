package rabbitmqinit

import (
	"fmt"
	"github.com/eugene-ruby/xconnect/rabbitmq"
)

// DeclareExchanges ensures all necessary exchanges exist
func DeclareExchanges(ch rabbitmq.Channel) error {
	if ch == nil {
		return fmt.Errorf("DeclareExchanges: channel is nil")
	}
	return ch.ExchangeDeclare(
		"murmapp",
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
}
