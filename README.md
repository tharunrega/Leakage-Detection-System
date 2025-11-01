# StackGuard Leakage Detection System (Go + REST)

A production-style microservice that scans GitHub public code for leaked tokens from your `inventory.json`, enriches results with user geolocation (GitHub profile), and sends alerts via Slack and Email.

## Features
- GitHub Code Search API (exact token search with quotes)
- Token inventory (`inventory.json`)
- Geolocation enrichment (GitHub profile `location`)
- Slack & Email alerts
- REST APIs: `/scan`, `/results`, `/alerts/test`, `/alerts/email/test`, `/scan/status`, `/health`
- Dockerized

## Endpoints
- `POST /scan` – starts a scan
- `GET /results` – returns latest results
- `POST /alerts/test` – sends a test Slack alert
- `POST /alerts/email/test` – sends a test email via configured SMTP (Mailtrap recommended)
- `GET /scan/status` – returns whether a scan is currently running
- `GET /health` – liveness check

## Setup

### 1) Clone & Env
```
cp .env.example .env
# fill in GITHUB_TOKEN, SLACK_WEBHOOK_URL, SMTP_* vars
```

### 2) Inventory
Edit `inventory.json` with token entries (type/value/owner). Do NOT commit real tokens — use `inventory.example.json` for examples.

### 3) Run locally
```
go run main.go
# in another terminal
curl -X POST http://localhost:8080/scan
curl http://localhost:8080/results
```

### 4) Run with Docker
```
docker build -t stackguard .
docker run --env-file .env -p 8080:8080 stackguard
```

## Security / Notes
- Do NOT commit real secrets (GITHUB_TOKEN, SLACK_WEBHOOK_URL, SMTP_*) to version control. If you accidentally committed them, remove the files from git and rotate the credentials immediately.
- Use `.env.example` and `inventory.example.json` as templates for local configuration.
- Add a `.dockerignore` to exclude `.env` and `inventory.json` from Docker builds to avoid baking secrets into images.
- The GitHub Search API returns file metadata; for full context/snippet the service optionally fetches file contents using the Contents API.
