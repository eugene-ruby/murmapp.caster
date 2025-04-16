# 📨 murmapp.caster

**murmapp.caster** is a lightweight message delivery service for the Murmapp platform.  
It listens to RabbitMQ messages with routing key `telegram.send`, decodes a protobuf message containing raw Telegram API JSON, and forwards it to the configured Telegram Bot API endpoint.

---

## 🧠 Overview

<p><img src="docs/img.jpg" /></p>

This service:

- Connects to RabbitMQ (`murmapp.messages.out` exchange)
- Listens to the `telegram.send` routing key
- Receives `SendMessageRequest` protobuf messages:
  ```proto
  message SendMessageRequest {
    string bot_api_key = 1;
    string api_endpoint = 2;
    bytes raw_body = 3;
  }
  ```
- Makes HTTP requests to `${TELEGRAM_API_URL}/${api_endpoint}`
- Sends raw JSON as the request body, and uses `bot_api_key` as Bearer token

---

## 📦 Installation

```bash
git clone https://github.com/eugene-ruby/murmapp.caster.git
cd murmapp.caster
go mod tidy
protoc --go_out=. --go_opt=paths=source_relative proto/send_message.proto
go build -o app ./cmd/main.go
```

Or use Docker:

```bash
docker build -f Dockerfile.dev -t murmapp-caster .
```

---

## ⚙️ Environment variables

| Variable            | Description                                  |
|---------------------|----------------------------------------------|
| `RABBITMQ_URL`      | RabbitMQ connection URL (e.g. `amqp://...`) |
| `TELEGRAM_API_URL`  | Base URL of Telegram API                    |

You can use `.env`:

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
TELEGRAM_API_URL=https://api.telegram.org
```

---

## 🧪 Running

```bash
./app
```

If all is well, you will see:

```
caster is running...
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

A sample unit test is provided in `internal/consumer_test.go`  
Run tests with:

```bash
go test ./internal
```

---

## 📋 License

MIT or custom license.
