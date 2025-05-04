package config

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/eugene-ruby/xencryptor/xsecrets"
)

// MasterEncryptionKey is the master secret key injected at build time via -ldflags.
// It is used for decrypting sensitive data like private RSA keys.
var MasterEncryptionKey string

func MasterKeyBytes() []byte {
	return []byte(MasterEncryptionKey)
}

// Config holds all configuration for the application.
type Config struct {
	AppPort     string
	SecretSalt  string
	WebhookHost string
	TelegramAPI string
	MasterKey   string
	RabbitMQ    RabbitMQConfig
	Redis       RedisConfig
	Encryption  EncryptionConfig
}

type RabbitMQConfig struct {
	URL string
}
type RedisConfig struct {
	URL string
}

type EncryptionConfig struct {
	PayloadEncryptionKeyStr    string
	SecretBotEncryptionKeyStr  string
	PrivateRSAEncryptionKeyStr string
	PayloadEncryptionKey       []byte
	SecretBotEncryptionKey     []byte
	PrivateRSAEncryptionKey    *rsa.PrivateKey
}

type defaultENV struct {
	appPort string
}

// LoadConfig reads environment variables and returns a Config instance.
func LoadConfig() (*Config, error) {
	defaultValues := &defaultENV{
		appPort: "8080",
	}

	cfg := &Config{
		AppPort:     os.Getenv("APP_PORT"),
		SecretSalt:  os.Getenv("SECRET_SALT"),
		WebhookHost: os.Getenv("WEB_HOOK_HOST"),
		TelegramAPI: os.Getenv("TELEGRAM_API_URL"),
		RabbitMQ: RabbitMQConfig{
			URL: os.Getenv("RABBITMQ_URL"),
		},
		Redis: RedisConfig{
			URL: os.Getenv("REDIS_URL"),
		},
		Encryption: EncryptionConfig{
			PayloadEncryptionKeyStr:    os.Getenv("PAYLOAD_ENCRYPTION_KEY"),
			SecretBotEncryptionKeyStr:  os.Getenv("SECRET_BOT_ENCRYPTION_KEY"),
			PrivateRSAEncryptionKeyStr: os.Getenv("ENCRYPTED_PRIVATE_KEY"),
		},
	}

	if cfg.SecretSalt == "" || len(cfg.SecretSalt) < 8 {
		return nil, fmt.Errorf("SECRET_SALT environment variable must be set and at least 8 characters long")
	}
	if cfg.WebhookHost == "" {
		return nil, fmt.Errorf("WEB_HOOK_HOST environment variable must be set")
	}
	if cfg.TelegramAPI == "" {
		return nil, fmt.Errorf("TELEGRAM_API_URL environment variable must be set")
	}
	if cfg.RabbitMQ.URL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL environment variable must be set")
	}
	if cfg.Redis.URL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable must be set")
	}
	if cfg.Encryption.PayloadEncryptionKeyStr == "" {
		return nil, fmt.Errorf("PAYLOAD_ENCRYPTION_KEY environment variable must be set")
	}
	if cfg.Encryption.SecretBotEncryptionKeyStr == "" {
		return nil, fmt.Errorf("SECRET_BOT_ENCRYPTION_KEY environment variable must be set")
	}
	if cfg.Encryption.PrivateRSAEncryptionKeyStr == "" {
		return nil, fmt.Errorf("ENCRYPTED_PRIVATE_KEY environment variable must be set")
	}
	if cfg.AppPort == "" {
		cfg.AppPort = defaultValues.appPort
	}

	if MasterEncryptionKey == "" {
		return nil, fmt.Errorf("MasterEncryptionKey must be injected at build time with -ldflags")
	}

	if err := decryptEncryptionKeys(cfg); err != nil {
		return nil, err
	}
	if err := decryptPrivateKey(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func decryptEncryptionKeys(cfg *Config) error {
	keyPayload := xsecrets.DeriveKey(MasterKeyBytes(), "payload")
	decryptedPayloadKey, err := xsecrets.DecryptBase64WithKey(cfg.Encryption.PayloadEncryptionKeyStr, keyPayload)

	if err != nil {
		return fmt.Errorf("failed to decrypt PAYLOAD_ENCRYPTION_KEY: %w", err)
	}
	cfg.Encryption.PayloadEncryptionKey = decryptedPayloadKey

	keyBot := xsecrets.DeriveKey(MasterKeyBytes(), "bot")
	decryptedSecretBotKey, err := xsecrets.DecryptBase64WithKey(cfg.Encryption.SecretBotEncryptionKeyStr, keyBot)
	if err != nil {
		return fmt.Errorf("failed to decrypt SECRET_BOT_ENCRYPTION_KEY: %w", err)
	}
	cfg.Encryption.SecretBotEncryptionKey = decryptedSecretBotKey

	return nil
}

func decryptPrivateKey(cfg *Config) error {
	encRSABase64 := cfg.Encryption.PrivateRSAEncryptionKeyStr
	masterKey := MasterKeyBytes()
	privateKey, err := xsecrets.DecryptPrivateRSA(encRSABase64, string(masterKey), "privateKey")

	if err != nil {
		return fmt.Errorf("failed to decrypt PrivateRSA: %w", err)
	}
	cfg.Encryption.PrivateRSAEncryptionKey = privateKey
	return nil
}
