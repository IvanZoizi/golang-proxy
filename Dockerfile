# ==================== Builder Stage ====================
FROM golang:1.25.4 AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение из правильной папки
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" \
    -o main ./cmd/main

# ==================== Final Stage ====================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && adduser --disabled-password --gecos "" appuser

WORKDIR /app

COPY --from=builder /app/main .

RUN chown appuser:appuser main && chmod +x main

USER appuser

EXPOSE 8080

CMD ["./main"]