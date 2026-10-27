# golang-proxy

> A configurable HTTP proxy written in Go, built on clean architecture — with IP access lists, rate limiting, caching, Prometheus metrics and Swagger docs.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=flat-square&logo=gin&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=flat-square&logo=prometheus&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-85EA2D?style=flat-square&logo=swagger&logoColor=black)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)

## About

The service runs two HTTP servers side by side:

- a **proxy server** (default `:8000`) that forwards traffic while applying access control, rate limiting and caching;
- a **management API** (default `:8080`, Gin) to manage the lists, inspect the cache and expose metrics and docs.

## Features

- 🛡️ **IP access lists** — white / gray / black lists, persisted as JSON and editable at runtime
- 🚦 **Rate limiting** — per-client limits backed by a dedicated repository
- ⚡ **Response cache** — with stats, config, invalidation and clear operations
- 📈 **Prometheus metrics** + ready-to-use **Grafana** dashboards (`config/grafana`)
- 📚 **Swagger** UI for the management API
- 🪵 Structured logging with **zap** and graceful shutdown
- 🐳 **Docker** / `docker-compose` for the full stack (proxy + Prometheus + Grafana)

## Architecture

Clean architecture, layered by responsibility:

```
cmd/main            # entrypoint & wiring
internal/
├── transport/http  # Gin handlers (IP lists, cache, health)
├── usecase         # business logic (lists, rate limiter, cache)
├── repository      # JSON-backed storage
├── entity          # domain models
├── dto             # request/response objects
└── metrics         # Prometheus collectors
pkg/logger          # zap setup
configGolang        # app & cache config loading
```

## Management API

| Route | Purpose |
|---|---|
| `GET /health` | Health check |
| `GET /metrics` | Prometheus metrics |
| `GET /swagger/index.html` | API documentation |
| `… /api/white`, `/api/gray`, `/api/black` | Manage IP lists |
| `GET /api/white/check` | Check whether an IP is listed |
| `… /api/cache/stats`, `/cache/config`, `/cache/invalidate`, `/cache/clear` | Inspect & manage the cache |

Authorization uses a Bearer token (`Authorization: Bearer <token>`).

## Configuration

`config.json`:

```json
{
  "serverPort": ":8080",
  "proxyServerPort": ":8000",
  "appName": "Proxy",
  "env": "development",
  "secretKey": "secret-key"
}
```

IP lists and rate-limiter state live in `data/*.json` (`whitelist.json`, `graylist.json`, `blacklist.json`, `rateLimiter.json`).

## Getting started

```bash
# with Docker (proxy + Prometheus + Grafana)
docker compose up --build

# or locally
go mod download
go run ./cmd/main
```

Swagger: http://localhost:8080/swagger/index.html · Metrics: http://localhost:8080/metrics

## Tech stack

**Go · Gin · Prometheus · Grafana · Swagger (swaggo) · zap · Docker**
