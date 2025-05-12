# 📩 About murmapp.caster v0.1.X

```bash
╭────────────────────────────────────────────────╮
│     murmapp.caster: trusted boundary for       │
│     decrypting & dispatching Telegram ops      │
╰────────────────────────────────────────────────╯
       ↳ RSA/AES-secured, event-driven, minimal
```


[![Go Report Card](https://goreportcard.com/badge/github.com/eugene-ruby/murmapp.caster)](https://goreportcard.com/report/github.com/eugene-ruby/murmapp.caster)
[![Build Status](https://github.com/eugene-ruby/murmapp.caster/actions/workflows/ci.yml/badge.svg)](https://github.com/eugene-ruby/murmapp.caster/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 🔗 Part of the Murmapp Ecosystem

Murmapp is composed of several specialized microservices:

| Project           | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| [`murmapp.hook`](https://github.com/eugene-ruby/murmapp.hook)     | Secure Telegram webhook receiver with ID redaction and encryption |
| [`murmapp.caster`](https://github.com/eugene-ruby/murmapp.caster) | Trusted dispatcher for decrypted message delivery to Telegram API |
| [`murmapp.core`](https://github.com/eugene-ruby/murmapp.core)     | Domain logic and message orchestration layer                       |
| [`murmapp.docs`](https://github.com/eugene-ruby/murmapp.docs)     | You are here — full technical documentation and architecture        |

> All components are open-source and licensed under MIT.

---

**murmapp.caster** is a secure and minimal microservice in the Murmapp ecosystem responsible for sending encrypted messages and registering webhooks with the Telegram Bot API.

It is the only component that holds decrypted bot_api_key credentials and has permission to contact Telegram directly. All incoming data is encrypted, processed securely, and passed through RabbitMQ.

In the Murmapp architecture, raw telegram_id values are never transmitted in plaintext. Instead, they are immediately encrypted in murmapp.hook using the public RSA key of murmapp.caster. Across the system, only a derived salted hash (XID) is used for indexing and routing.

Upon receiving an encrypted Telegram ID, murmapp.caster decrypts it with its private RSA key and then re-encrypts the value using AES for efficient future access and lookup.

---

## 🔧 Features

* Listens to RabbitMQ messages and dispatches them to Telegram
* Handles Telegram webhook registration securely
* Performs encrypted Telegram ID resolution via Redis/PostgreSQL
* Ensures no sensitive data is stored in plaintext
* Includes health check server on `/healthz`

---

## 🚀 Quick Start

A template file env_test_example is provided for development and testing purposes. It included in Makefile via:

```bash
include .env_test
export
```

To use it, simply rename the template:

```bash
mv env_test_example .env_test
```

and adjust the values to match your environment.

```bash
make build
```

---

## 📁 Environment Variables

| Variable                 | Required | Description                                                 |
| ------------------------ | -------- | ----------------------------------------------------------- |
| `APP_PORT`               | No       | Port for health check endpoint (default `8080`)             |
| `WEB_HOOK_HOST`          | Yes      | Base URL for webhook registration                           |
| `TELEGRAM_API_URL`       | Yes      | URL to Telegram Bot API                                     |
| `RABBITMQ_URL`           | Yes      | Connection URI to RabbitMQ                                  |
| `REDIS_URL`              | Yes      | Redis server address                                        |
| `POSTGRES_DSN`           | Yes      | PostgreSQL DSN used by storewriter                          |
| `SECRET_SALT`            | Yes      | Base64-encoded and encrypted salt (for webhook/XID)         |
| `PAYLOAD_ENCRYPTION_KEY` | Yes      | Base64-encoded and encrypted AES key for payload decryption |
| `ENCRYPTED_PRIVATE_KEY`  | Yes      | Base64-encoded and encrypted PEM-encoded RSA private key    |
| `PUBLIC_KEY_RAW_BASE64`  | Optional | Base64-encoded raw RSA public key (used in tests)           |

> All encrypted values are decrypted at runtime using `MasterEncryptionKey` injected via `-ldflags`.

---

## 🔄 Message Flows


### ✉️ SendMessage Flow

```bash
                        ┌────────────────────────────┐
                        │  telegram.messages.in      │
                        │  (TelegramWebhookPayload)  │
                        └────────────┬───────────────┘
                                     ▼
                           ┌───────────────────┐
                           │  caster (Go svc)  │
                           │  decrypt payload  │
                           └────────┬──────────┘
                                    ▼
             ┌────────────────────────────┐
             │  detect {XID:...} patterns │
             │  fetch original IDs from   │
             │  Redis via XID             │
             └──────────┬─────────────────┘
                        ▼
             ┌────────────────────────────┐
             │ inject real telegram_id    │
             │ into payload               │
             └──────────┬─────────────────┘
                        ▼
             ┌────────────────────────────┐
             │   send to Telegram API     │
             └────────────────────────────┘


                        ┌────────────────────────────┐
                        │ telegram.encrypted.id      │
                        │ (EncryptedTelegramID)      │
                        └────────────┬───────────────┘
                                     ▼
                           ┌───────────────────┐
                           │     decrypt RSA   │
                           │     store in DB   │
                           └───────────────────┘

```

* Queue: `telegram.messages.out`
* Format: `SendMessageRequest` (see `proto/send_message.proto`)
* Steps:

  * Decrypts `bot_api_key` and `payload`
  * Resolves `__XID:<hash>__` placeholders to real `telegram_id`
  * Sends raw JSON to Telegram API

### 🔑 Webhook Registration Flow

```bash
            ┌────────────────────────────┐
            │   webhook.registration     │
            │   (RegisterWebhookRequest) │
            └────────────┬───────────────┘
                         ▼
              ┌──────────────────────┐
              │  caster (Go svc)     │
              │  decrypt bot API key │
              └────────┬─────────────┘
                       ▼
       ┌─────────────────────────────────────────┐
       │ Generate:                               │
       │   - webhook_id = SHA256(secret + salt)  │
       │   - secret_token                        │
       │   - callback URL                        │
       └────────────┬────────────────────────────┘
                    ▼
       ┌─────────────────────────────────────────┐
       │   Send setWebhook to Telegram API       │
       │   (with url + secret_token)             │
       └────────────┬────────────────────────────┘
                    ▼
       ┌─────────────────────────────────────────┐
       │  Publish RegisterWebhookResponse        │
       │  to webhook.registered                  │
       └─────────────────────────────────────────┘

```

* Queue: `webhook.registration`
* Format: `RegisterWebhookRequest` (see `proto/registration.proto`)
* Steps:

  * Decrypts API key
  * Generates secure `secret_token`
  * Registers webhook via Telegram
  * Publishes `RegisterWebhookResponse` to `webhook.registered`

### 🚪 Telegram ID Storage

* Queue: `telegram.encrypted.id`
* Format: `EncryptedTelegramID` (see `proto/encrypted_telegram_id.proto`)
* Steps:

  * Decrypts Telegram ID using private RSA key
  * Re-encrypts with AES and stores in PostgreSQL `telegram_id_map`
  * Skips if XID already exists

---

## 🖇️ Schema Setup

```bash
psql < setup/init_db.sql
psql < setup/migrate.sql
```

This initializes:

* `telegram_id_map` — for XID to Telegram ID storage
* Necessary indexes for fast resolution

---

## 🛡️ Security Model

* No plaintext secrets stored or transmitted
* All data passed via Protobuf with encrypted fields (AES-GCM)
* `telegram_id` stored encrypted; never exposed in logs
* `bot_api_key` never written to disk
* Master key never included in `.env`, passed via build-time flag

---

## ✨ Health Check

```bash
curl http://localhost:$APP_PORT/healthz
```

Returns `ok`.

---

## ✅ Testing

```bash
rename env_test_example .env_test
make test
```

Includes unit tests for:

* Telegram API call logic
* Registration logic
* Redis+PostgreSQL XID resolution
* Full encryption/decryption lifecycle

---

## 🌍 First Public Release — v0.1.1

This is the first public stable release of `murmapp.caster`. Feedback and contributions are welcome.

---

## ™ License

This project is licensed under the [MIT License](/LICENSE).
