package telegramout

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	casterpb "murmappcaster/proto"

	"github.com/eugene-ruby/xconnect/redisstore"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type XIDResolver struct {
	Redis                   *redisstore.Store
	DB                      *sql.DB
	TelegramIdEncryptionKey []byte
	TTL                     time.Duration
}

func (r *XIDResolver) Resolve(ctx context.Context, xid string) ([]byte, error) {
	data, err := r.Redis.Get(ctx, xid)
	if err == nil {
		return r.decryptFromBytes(data, xid)
	}
	if err != redis.Nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var encrypted []byte
	err = r.DB.QueryRowContext(ctx, `
		SELECT encrypted_id FROM telegram_id_map WHERE telegram_xid = $1
	`, xid).Scan(&encrypted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pg select failed: %w", err)
	}

	record := &casterpb.TelegramIdStore{
		Version:          "v1",
		EncryptedPayload: encrypted,
	}
	raw, err := proto.Marshal(record)
	if err == nil {
		_ = r.Redis.Set(ctx, xid, string(raw), r.TTL)
	}

	return r.decryptFromProto(record, xid)
}

func (r *XIDResolver) decryptFromBytes(data string, xid string) ([]byte, error) {
	var record casterpb.TelegramIdStore
	if err := proto.Unmarshal([]byte(data), &record); err != nil {
		return nil, fmt.Errorf("unmarshal redis[%s]: %w", xid, err)
	}
	return r.decryptFromProto(&record, xid)
}

func (r *XIDResolver) decryptFromProto(record *casterpb.TelegramIdStore, xid string) ([]byte, error) {
	if record.Version != "v1" {
		return nil, nil
	}
	decrypted, err := xsecrets.DecryptBytesWithKey(record.EncryptedPayload, r.TelegramIdEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("aes decrypt xid[%s]: %w", xid, err)
	}
	return decrypted, nil
}
