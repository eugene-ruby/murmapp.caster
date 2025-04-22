# Changelog

All notable changes to **murmapp.caster** will be documented in this file.

---

## [v0.1.0] - 2025-04-22

### ✨ Initial Public Release

- First stable version of `murmapp.caster`
- Secure RabbitMQ-based message handler for Telegram Bot API

### ✅ Features

- AES-GCM encryption for:
  - `bot_api_key`
  - `telegram_id`
  - `payload` (Telegram message JSON)
- Subscribes to:
  - `telegram.messages.out` → handles message sending
  - `webhook.registration` → performs webhook registration
- Pushes confirmation to `webhook.registered`
- Only component with access to decrypted `bot_api_key`
- Clean HTTP `/healthz` endpoint
- CI-ready, containerized
- ~77% code coverage with full unit test support

---

🛡 Security-first design.  
🚀 Ready for production.
