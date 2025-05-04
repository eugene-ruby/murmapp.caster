include .env_test
export

# Fully qualified Go symbol (matches -ldflags -X path)
MASTER_KEY_VAR := murmapp.caster/internal/config.MasterEncryptionKey

# Run tests with injected master key via -ldflags
test:
	go test -v ./internal/config -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"
	POSTGRES_DSN=$(POSTGRES_DSN) go test -v ./internal/run_test.go -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"
	go test -v ./internal/registration
	go test -v ./internal/telegramout
	POSTGRES_DSN=$(POSTGRES_DSN) go test -v ./internal/storewriter/...

build:
	go build -ldflags "-X=murmapp.caster/internal/config.MasterEncryptionKey=$(MASTER_KEY)" -o casterapp ./cmd/main.go