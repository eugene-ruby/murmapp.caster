package storewriter

import (
	"context"
	"database/sql"
	"log"

	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
)

type Handler struct {
	DB *sql.DB
}

func HandleEncryptedID(body []byte, h *Handler) {
	var msg casterpb.EncryptedTelegramID
	if err := proto.Unmarshal(body, &msg); err != nil {
		log.Printf("[storewriter] ❌ failed to decode proto: %v", err)
		return
	}

	ctx := context.Background()
	_, err := h.DB.ExecContext(ctx, `
		INSERT INTO telegram_id_map (telegram_xid, encrypted_id, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (telegram_xid) DO NOTHING
	`, msg.TelegramXid, msg.EncryptedId)
	if err != nil {
		log.Printf("[storewriter] ❌ failed to insert: %v", err)
	}
}
