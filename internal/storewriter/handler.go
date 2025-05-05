package storewriter

import (
	"context"
	"database/sql"
	"log"
	"crypto/rsa"

	"github.com/eugene-ruby/xencryptor/xsecrets"
	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

type Handler struct {
	DB *sql.DB
	MasterKey []byte
	PrivateKey *rsa.PrivateKey
}

func HandleEncryptedID(body []byte, h *Handler) {
	var msg casterpb.EncryptedTelegramID
	if err := proto.Unmarshal(body, &msg); err != nil {
		log.Printf("[storewriter] ❌ failed to decode proto: %v", err)
		return
	}

	decrypt_id, err := xsecrets.RSADecryptBytes(msg.EncryptedId, h.PrivateKey)
	if err != nil {
		log.Printf("[storewriter] ❌ failed to decrypted telegram_id: %v", err)
		return
	}

	key := xsecrets.DeriveKey(h.MasterKey, "telegram_id")
	encrypted_id, err := xsecrets.EncryptBase64WithKey(decrypt_id, key)
	if err != nil {
		log.Printf("[storewriter] ❌ failed to encrypted telegram_id: %v", err)
		return
	}

	ctx := context.Background()
	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO telegram_id_map (telegram_xid, encrypted_id, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (telegram_xid) DO NOTHING
	`, msg.TelegramXid, encrypted_id)
	if err != nil {
		log.Printf("[storewriter] ❌ failed to insert: %v", err)
	}
}
