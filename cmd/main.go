package main

import (
    "log"

    "murmapp.caster/internal"
)

func main() {
    if err := internal.InitEnv(); err != nil {
		log.Fatalf("ENV init failed: %v", err)
	}

    mq, err := internal.InitMQ()
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer mq.Close()

	if err := internal.InitExchanges(mq.GetChannel()); err != nil {
		log.Fatalf("Exchange init failed: %v", err)
	}

    if err := internal.InitEncryptionKey(); err != nil {
		log.Fatalf("Encryption key init failed: %v", err)
	}

	go func() {
		if err := internal.StartRegistrationConsumer(mq.GetChannel()); err != nil {
			log.Fatalf("failed to start registration consumer: %v", err)
		}
	}()

	go func() {
		if err := internal.StartConsumerMsgOUT(mq.GetChannel(), internal.TelegramAPI); err != nil {
			log.Fatalf("failed to start registration consumer message out: %v", err)
		}
	}()

	select {}
}
