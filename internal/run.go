package internal

import (
    "log"
	"net/http"
	"os"
	"fmt"

	"github.com/go-chi/chi/v5"

)

var InitMQFunc = InitMQ
var InitExchangesFunc = InitExchanges
var StartRegistrationConsumerFunc = StartRegistrationConsumer
var StartConsumerMsgOutFunc = StartConsumerMsgOut

func Run() error {
	// init env vars
	if err := InitEncryptionKey(); err != nil {
		return fmt.Errorf("encryption key init failed: %w", err)
	}
	if err := InitEnv(); err != nil {
		return fmt.Errorf("env init failed: %w", err)
	}
	if err := InitPrivateRSA(); err != nil {
		return fmt.Errorf("encryption rsa key init failed: %w", err)
	}

	mq, err := InitMQFunc()
	if err != nil {
		return fmt.Errorf("rabbitmq init failed: %w", err)
	}
	defer mq.Close()

	if err := InitExchangesFunc(mq); err != nil {
		return fmt.Errorf("exchange init failed: %w", err)
	}

	// start consumers
	go func() {
		if err := StartRegistrationConsumerFunc(mq); err != nil {
			log.Fatalf("failed to start registration consumer: %v", err)
		}
	}()

	go func() {
		if err := StartConsumerMsgOutFunc(mq, TelegramAPI); err != nil {
			log.Fatalf("failed to start message out consumer: %v", err)
		}
	}()

	// HTTP server for healthz
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3005"
	}
	addr := ":" + port

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Printf("🌐 Starting server on %s...", addr)
	return http.ListenAndServe(addr, r)
}
