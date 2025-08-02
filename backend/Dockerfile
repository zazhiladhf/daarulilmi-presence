# Tahap 1: Build binary dari Golang
FROM --platform=linux/arm64 golang:1.21-alpine AS builder

# Direktori kerja
WORKDIR /app

# Salin go.mod dan go.sum untuk unduh dependen awal
COPY go.mod go.sum ./
RUN go mod download

# Salin seluruh source code (termasuk cmd/)
COPY . .

# Build binary dari path khusus
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o app ./cmd/web

# Tahap 2: Runtime
FROM --platform=linux/arm64 alpine:latest

WORKDIR /app

# Salin hasil build dari tahap 1
COPY --from=builder /app/app .

# Jalankan binary
CMD ["./app"]
