package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/eugene-ruby/xencryptor/xsecrets"
)

func TestLoadConfig_Success(t *testing.T) {
	// Set up environment variables for a successful configuration load
	os.Setenv("SECRET_SALT", "somesupersecretsalt")
	os.Setenv("WEB_HOOK_HOST", "https://example.com")
	os.Setenv("TELEGRAM_API_URL", "https://api.telegram.org")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	os.Setenv("REDIS_URL", "localhost:6379")
	os.Setenv("POSTGRES_DSN", "postgres://user@localhost:5432/base")

	// Set the build-time injected master encryption key manually
	masterEncryptionKey := "test-master-key"

	payloadKey := "payloadkey12345678"
	payloadMasterKey := xsecrets.DeriveKey([]byte(masterEncryptionKey), "payload")
	payloadEncryptKey, _ := xsecrets.EncryptBase64WithKey([]byte(payloadKey), payloadMasterKey)
	os.Setenv("PAYLOAD_ENCRYPTION_KEY", payloadEncryptKey)

	secretSaltKey := "secretsalt123456"
	secretMasterKey := xsecrets.DeriveKey([]byte(masterEncryptionKey), "salt")
	secretSaltEncryptKey, _ := xsecrets.EncryptBase64WithKey([]byte(secretSaltKey), secretMasterKey)
	os.Setenv("SECRET_SALT", secretSaltEncryptKey)

	pemPrivateBytes, _, _ := xsecrets.GenerateKeyPair()
	cipherText, _ := xsecrets.EncryptPrivateRSA(pemPrivateBytes, masterEncryptionKey, "privateKey")
	os.Setenv("ENCRYPTED_PRIVATE_KEY", cipherText)

	secretBotKey := xsecrets.DeriveKey([]byte(masterEncryptionKey), "bot")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "https://example.com", cfg.WebhookHost)
	require.Equal(t, "https://api.telegram.org", cfg.TelegramAPI)
	require.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQ.URL)
	require.Equal(t, []byte(payloadKey), cfg.Encryption.PayloadEncryptionKey)
	require.Equal(t, secretBotKey, cfg.Encryption.SecretBotEncryptionKey)
	require.Equal(t, []byte("secretsalt123456"), cfg.Encryption.SecretSalt)
}

func TestLoadConfig_MissingVariables(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	// Clear the build-time master encryption key
	MasterEncryptionKey = ""

	_, err := LoadConfig()
	require.Error(t, err)
}
