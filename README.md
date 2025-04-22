# 📨 murmapp.caster

**murmapp.caster** is a secure microservice in the Murmapp ecosystem responsible for handling all communication with the **Telegram Bot API**.

It is the **only component** with access to decrypted bot credentials (`bot_api_key`) and is trusted to send messages and register webhooks on behalf of bots.

All communication is done via **RabbitMQ**, with encrypted Protobuf payloads that include:

- `telegram_id` — AES-GCM encrypted (shared with `hook`)
- `bot_api_key` — AES-GCM encrypted (exclusive to `caster`)
- `raw Telegram payload` — AES-GCM encrypted JSON (shared with `core` and `hook`)

`caster` receives these encrypted commands and securely delivers them to the Telegram API.

---

## 🔬 Overview
<p><img src="docs/context_c4diagram.jpg"></p>

This service:

- Connects to RabbitMQ (`murmapp` topic exchange)
- Listens to two routing keys:
  - `telegram.messages.out` — for sending messages
  - `webhook.registration` — for registering Telegram bot webhooks
- Receives encrypted Protobuf messages:
  
```proto
    message SendMessageRequest {
        bytes encrypted_api_key_bot = 1;
        string api_endpoint = 2;
        bytes encrypted_payload = 3;
    }

    message RegisterWebhookRequest {
        string bot_id = 1;
        bytes api_key_bot = 2;
    }

    message RegisterWebhookResponse {
        string bot_id = 1;
        bytes encrypted_api_key_bot = 2;
        string webhook_id = 3;
    }
```
 - Decrypts sensitive fields using AES-GCM and local keys
 - Sends HTTP requests to `${TELEGRAM_API_URL}/bot${api_key}/${api_endpoint}`
 - Uses decrypted api_key_bot as part of the URL (Telegram format)
 - Sends decrypted JSON payloads as raw HTTP body
 - After successful webhook registration, pushes confirmation to webhook.registered

---

## 📦 Installation

```bash
git clone https://github.com/eugene-ruby/murmapp.caster.git
cd murmapp.caster
go mod tidy
protoc --go_out=. --go_opt=paths=source_relative proto/*.proto
go build -o app ./cmd/main.go
```

Or use Docker:

```bash
docker build -f Dockerfile.dev -t murmapp-caster .
```

---

## ⚙️ Environment variables
## ⚙️ Environment Variables

These environment variables must be defined to run the service:

| Variable                     | Description                                                                 |
|------------------------------|-----------------------------------------------------------------------------|
| `TELEGRAM_API_URL`           | Base URL for the Telegram Bot API (e.g. `https://api.telegram.org`)        |
| `BOT_ENCRYPTION_KEY`         | 32-byte AES key used to decrypt encrypted `bot_api_key`                    |
| `ENCRYPTION_KEY`             | 32-byte AES key for decrypting `payload` (shared with `hook`, `core`)     |
| `TELEGRAM_ID_ENCRYPTION_KEY`| 32-byte AES key for decrypting `telegram_id` (shared with `hook`)         |
| `RABBITMQ_URL`               | AMQP URI to connect to RabbitMQ (e.g. `amqp://guest:guest@localhost:5672`)|
| `WEB_HOOK_HOST`              | Base host used when registering webhooks (e.g. `https://api.myapp.com`)   |
| `SECRET_SALT`                | Salt used to generate unique webhook IDs                                   |
| `APP_PORT`                   | Port number for the HTTP server (`/healthz` endpoint)                      |

---

### 📄 Example `.env` file

```dotenv
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
TELEGRAM_API_URL=https://api.example.com
SECRET_SALT=salt12434
ENCRYPTION_KEY=01234567890123456789012345678901
TELEGRAM_ID_ENCRYPTION_KEY=12345678901234567890123456789012
BOT_ENCRYPTION_KEY=12345678901234567890123456789069
WEB_HOOK_HOST=https://myaip.example.com/api/webhook/
APP_PORT=8080
```

---

## 🧪 Running

```bash
dotenv run -- ./app
```

If all is well, you will see:

```
2025/04/22 11:45:38 ✅ Connected to RabbitMQ on attempt 1
2025/04/22 11:45:38 exchange declared: murmapp
2025/04/22 11:45:38 🌐 Starting server on :8080...
2025/04/22 11:45:38 [caster.registrations] 📖 consumer started and listening...
2025/04/22 11:45:38 🗣️  caster is running...
```

---

## 🛠 Systemd service (production)

To run `murmapp.caster` as a native service:

📄 `/etc/systemd/system/murmapp-caster.service`:

```ini
[Unit]
Description=Murmapp Caster Service
After=network.target

[Service]
User=deploy
WorkingDirectory=/home/deploy/murmapp/murmapp.caster
ExecStart=/home/deploy/murmapp/murmapp.caster/app
Restart=always
RestartSec=3
EnvironmentFile=/home/deploy/murmapp/murmapp.caster/.env

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl daemon-reexec
sudo systemctl enable murmapp-caster
sudo systemctl start murmapp-caster
```

---

## 🔁 Deployment via GitHub Actions

See `.github/workflows/deploy.yml`. It:

- Builds binary on GitHub
- Uploads to the server via `scp`
- Creates `.env` with secrets
- Restarts the `murmapp-caster` systemd service

---

## ✅ Test

Run tests with:

```bash
go test ./internal
```

---

## 📋 License

MIT or custom license.