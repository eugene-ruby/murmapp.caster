package storewriter_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
	"murmapp.caster/internal/storewriter"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var testDSN = os.Getenv("POSTGRES_DSN")

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", testDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`DELETE FROM telegram_id_map`)
	require.NoError(t, err)

	return db
}

func Test_HandleEncryptedID_insert_success(t *testing.T) {
	db := setupTestDB(t)
	handler := &storewriter.Handler{DB: db}

	msg := &casterpb.EncryptedTelegramID{
		TelegramXid: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd",
		EncryptedId: []byte("encrypted_data_here"),
	}
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)

	storewriter.HandleEncryptedID(raw, handler)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM telegram_id_map WHERE telegram_xid = $1`, msg.TelegramXid).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func Test_HandleEncryptedID_duplicate(t *testing.T) {
	db := setupTestDB(t)
	handler := &storewriter.Handler{DB: db}

	xid := "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef"

	_, err := db.Exec(`
		INSERT INTO telegram_id_map (telegram_xid, encrypted_id, created_at)
		VALUES ($1, $2, $3)
	`, xid, []byte("existing_data"), time.Now())
	require.NoError(t, err)

	msg := &casterpb.EncryptedTelegramID{
		TelegramXid: xid,
		EncryptedId: []byte("new_data_should_be_ignored"),
	}
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)

	storewriter.HandleEncryptedID(raw, handler)

	var data []byte
	err = db.QueryRow(`SELECT encrypted_id FROM telegram_id_map WHERE telegram_xid = $1`, xid).Scan(&data)
	require.NoError(t, err)
	require.Equal(t, []byte("existing_data"), data)
}

func Test_HandleEncryptedID_invalid_proto(t *testing.T) {
	db := setupTestDB(t)
	handler := &storewriter.Handler{DB: db}

	storewriter.HandleEncryptedID([]byte("not a proto"), handler)

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM telegram_id_map`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
