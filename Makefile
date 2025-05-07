include .env_test
export

# Fully qualified Go symbol (matches -ldflags -X path)
MASTER_KEY_ VAR := github.com/eugene-ruby/murmapp.caster/internal/config.MasterEncryptionKey

# Run tests with injected master key via -ldflags
test:
	go test --timeout 30s -v ./internal/config -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"
	go test --timeout 60s -v ./internal/registration -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"
	go test --timeout 60s -v ./internal/telegramout -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"
	go test --timeout 60s -v ./internal/storewriter -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"

build:
	go build -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)" -o casterapp ./cmd/main.go