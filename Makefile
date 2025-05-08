include .env_test
export

# Fully qualified Go symbol (matches -ldflags -X path)
MASTER_KEY_VAR := murmapp.caster/internal/config.MasterEncryptionKey

# Run tests with injected master key via -ldflags
test:
	go test --timeout 30s -v ./internal/... -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"

build:
	go build -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)" -o casterapp ./cmd/main.go