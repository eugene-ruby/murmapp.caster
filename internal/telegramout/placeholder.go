package telegramout

import (
	"context"
	"time"
	"fmt"
	"regexp"
	"strings"
	"database/sql"

	"github.com/eugene-ruby/xconnect/redisstore"
)

var xidPattern = regexp.MustCompile(`__XID:([a-f0-9]{64})__`)

type XIDPlaceholders struct {
	Redis                   *redisstore.Store
	DB                      *sql.DB
	TelegramIdEncryptionKey []byte
	TTL                     time.Duration
}

// ReplaceXIDPlaceholders replaces all __XID:hash__ values in JSON payload with decrypted Telegram IDs.
func ReplaceXIDPlaceholders(jsonText []byte, x *XIDPlaceholders) ([]byte, error) {
	matches := xidPattern.FindAllSubmatch(jsonText, -1)
	if len(matches) == 0 {
		return jsonText, nil
	}

	updated := string(jsonText)

	for _, match := range matches {
		fullPlaceholder := string(match[0]) // e.g., "__XID:abc...__"
		hash_xid := string(match[1])            // e.g., "abc..."

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		xIDResolver := &XIDResolver{
			Redis: x.Redis,
			DB: x.DB,
			TelegramIdEncryptionKey: x.TelegramIdEncryptionKey,
			TTL: x.TTL,
		}
		decryptedID, err := xIDResolver.Resolve(ctx, hash_xid)
		cancel()
		if decryptedID == nil {
			continue // no such key, skip
		}
		if err != nil {
			return nil, fmt.Errorf("redis error for hash %s: %w", hash_xid, err)
		}

		// Replace all occurrences (just in case)
		updated = strings.ReplaceAll(updated, fullPlaceholder, string(decryptedID))
	}

	return []byte(updated), nil
}
