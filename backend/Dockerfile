# Tahap 1: Builder pakai Go 1.24
FROM --platform=linux/arm64 golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o app ./cmd/web

# Tahap 2: Runtime tetap pakai Alpine
FROM --platform=linux/arm64 alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]
