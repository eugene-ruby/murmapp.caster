package telegramout

import (
	"context"
	"time"
	"fmt"
	"regexp"
	"strings"
	"crypto/rsa"

	"google.golang.org/protobuf/proto"
	casterpb "murmapp.caster/proto"
	"github.com/redis/go-redis/v9"
	"github.com/eugene-ruby/xencryptor/xsecrets"
	"github.com/eugene-ruby/xconnect/redisstore"
)


var xidPattern = regexp.MustCompile(`__XID:([a-f0-9]{64})__`)

// ReplaceXIDPlaceholders replaces all __XID:hash__ values in JSON payload with decrypted Telegram IDs.
func ReplaceXIDPlaceholders(jsonText []byte, store *redisstore.Store, rsaKey *rsa.PrivateKey) ([]byte, error) {
	matches := xidPattern.FindAllSubmatch(jsonText, -1)
	if len(matches) == 0 {
		return jsonText, nil
	}

	updated := string(jsonText)

	for _, match := range matches {
		fullPlaceholder := string(match[0]) // e.g., "__XID:abc...__"
		hash := string(match[1])            // e.g., "abc..."

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		storeVal, err := store.Get(ctx, hash)
		cancel()
		if err == redis.Nil {
			continue // no such key, skip
		}
		if err != nil {
			return nil, fmt.Errorf("redis error for hash %s: %w", hash, err)
		}

		var record casterpb.TelegramIdStore
		if err := proto.Unmarshal([]byte(storeVal), &record); err != nil {
			return nil, fmt.Errorf("❌ Failed to decode proto: redis[%s] failed: %w", hash, err)
		}

		if record.Version != "v1" {
			continue // unsupported version, skip
		}

		decryptedID, err := xsecrets.RSADecryptBytes(record.EncryptedPayload, rsaKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt redis[%s] failed: %w", hash, err)
		}

		// Replace all occurrences (just in case)
		updated = strings.ReplaceAll(updated, fullPlaceholder, string(decryptedID))
	}

	return []byte(updated), nil
}
