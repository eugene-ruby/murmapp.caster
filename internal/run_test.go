package internal_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"murmappcaster/internal"

	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/stretchr/testify/require"
)

func TestRun_HealthzOK(t *testing.T) {
	// 🔐 Master key, такой же как передаётся через -ldflags
	masterKey := "test-master-key"

	// 🔐 Генерация зашифрованных ключей
	payloadKey := "payload-key-1234567890"
	secretKey := "secretbot-key-abcdef"

	// derive & encrypt
	payloadEncKey, _ := xsecrets.EncryptBase64WithKey([]byte(payloadKey), xsecrets.DeriveKey([]byte(masterKey), "payload"))
	secretEncKey, _ := xsecrets.EncryptBase64WithKey([]byte(secretKey), xsecrets.DeriveKey([]byte(masterKey), "bot"))

	// private RSA (PEM) → зашифровать
	pemPriv, _, _ := xsecrets.GenerateKeyPair()
	privEncrypted, _ := xsecrets.EncryptPrivateRSA(pemPriv, masterKey, "privateKey")

	// ✅ ENV для Run()
	os.Setenv("APP_PORT", "3999")
	os.Setenv("SECRET_SALT", "somesalt")
	os.Setenv("WEB_HOOK_HOST", "https://example.com")
	os.Setenv("TELEGRAM_API_URL", "https://api.telegram.org")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("PAYLOAD_ENCRYPTION_KEY", payloadEncKey)
	os.Setenv("SECRET_BOT_ENCRYPTION_KEY", secretEncKey)
	os.Setenv("ENCRYPTED_PRIVATE_KEY", privEncrypted)

	// ⏳ Стартуем Run()
	go func() {
		err := internal.Run()
		require.NoError(t, err)
	}()

	// ⏲ Ждём, пока сервер поднимется
	time.Sleep(1 * time.Second)

	// ✅ Проверяем /healthz
	resp, err := http.Get("http://localhost:3999/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
