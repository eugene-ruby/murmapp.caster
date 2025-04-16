package main

import (
    "log"
    "os"

    "murmapp.caster/internal"
)

func main() {
    rabbitURL := os.Getenv("RABBITMQ_URL")
    if rabbitURL == "" {
        log.Fatal("RABBITMQ_URL is not set")
    }

    telegramAPI := os.Getenv("TELEGRAM_API_URL")
    if telegramAPI == "" {
        log.Fatal("TELEGRAM_API_URL is not set")
    }

    if err := internal.StartConsumer(rabbitURL, telegramAPI); err != nil {
        log.Fatalf("failed to start consumer: %v", err)
    }
}
