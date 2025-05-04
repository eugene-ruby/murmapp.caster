# Default value if not provided
MASTER_KEY ?= test-master-key

# Fully qualified Go symbol (matches -ldflags -X path)
MASTER_KEY_VAR := murmapp.caster/internal/config.MasterEncryptionKey

# Run tests with injected master key via -ldflags
test-config:
	go test -v ./internal/config -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"

test:
	go test -v ./internal/... -ldflags "-X=$(MASTER_KEY_VAR)=$(MASTER_KEY)"

build:
	go build -ldflags "-X=murmapp.caster/internal/config.MasterEncryptionKey=$(MASTER_KEY)" -o casterapp ./cmd/main.go