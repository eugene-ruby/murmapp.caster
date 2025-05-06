package storewriter_test

import (
	"encoding/base64"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/eugene-ruby/xencryptor/xsecrets"
	"murmapp.caster/internal/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"murmapp.caster/internal/storewriter"
	casterpb "murmapp.caster/proto"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T, pgDNS string) *sql.DB {
	db, err := sql.Open("pgx", pgDNS)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`DELETE FROM telegram_id_map`)
	require.NoError(t, err)

	return db
}

func setupTestHandler(t *testing.T) *storewriter.Handler {
	cfg, _ := config.LoadConfig()

	db := setupTestDB(t, cfg.PostgreSQL.DSN)
	handler := &storewriter.Handler{
		DB: db,
		TelegramIdEncryptionKey: cfg.Encryption.TelegramIdEncryptionKey,
		PrivateKey: cfg.Encryption.PrivateRSAEncryptionKey,
	}

	return handler
}

func Test_HandleEncryptedID_insert_success(t *testing.T) {
	handler := setupTestHandler(t)

	telegram_id := "1234567890"
	encryptXID, err := xsecrets.RSAEncryptBytes(publicRSAKey(), []byte(telegram_id))
	require.NoError(t, err)

	msg := &casterpb.EncryptedTelegramID{
		TelegramXid: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd",
		EncryptedId: encryptXID,
	}
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)

	storewriter.HandleEncryptedID(raw, handler)

	var data []byte
	err = handler.DB.QueryRow(`SELECT encrypted_id FROM telegram_id_map WHERE telegram_xid = $1`, msg.TelegramXid).Scan(&data)
	require.NoError(t, err)
	original_id, err := xsecrets.DecryptBytesWithKey(data, handler.TelegramIdEncryptionKey)
	require.NoError(t, err)
	require.Equal(t, []byte(telegram_id), original_id)
}

func Test_HandleEncryptedID_duplicate(t *testing.T) {
	handler := setupTestHandler(t)

	telegram_id := "1234567890"
	encryptXID, _ := xsecrets.RSAEncryptBytes(publicRSAKey(), []byte(telegram_id))
	xid := "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef"

	_, err := handler.DB.Exec(`
		INSERT INTO telegram_id_map (telegram_xid, encrypted_id, created_at)
		VALUES ($1, $2, $3)
	`, xid, []byte("current_existing_data"), time.Now())
	require.NoError(t, err)

	msg := &casterpb.EncryptedTelegramID{
		TelegramXid: xid,
		EncryptedId: encryptXID,
	}
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)

	storewriter.HandleEncryptedID(raw, handler)

	var data []byte
	err = handler.DB.QueryRow(`SELECT encrypted_id FROM telegram_id_map WHERE telegram_xid = $1`, xid).Scan(&data)
	require.NoError(t, err)
	require.Equal(t, []byte("current_existing_data"), data)
}

func Test_HandleEncryptedID_invalid_proto(t *testing.T) {
	handler := setupTestHandler(t)

	storewriter.HandleEncryptedID([]byte("not a proto"), handler)

	var count int
	err := handler.DB.QueryRow(`SELECT COUNT(*) FROM telegram_id_map`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func publicRSAKey() *rsa.PublicKey {
	s := os.Getenv("PUBLIC_KEY_RAW_BASE64")
	derBytes, _ := base64.RawStdEncoding.DecodeString(s)
	pubKey, _ := x509.ParsePKIXPublicKey(derBytes)
	rsaKey, _ := pubKey.(*rsa.PublicKey)

	return rsaKey
}