# --- Tahap 1: Builder ---
# Kita gunakan image resmi Go sebagai "ruang perakitan" sementara.
FROM golang:1.24-alpine AS builder

# Tentukan direktori kerja di dalam kontainer
WORKDIR /app

# Salin file manajemen modul terlebih dahulu untuk caching
COPY go.mod ./
COPY go.sum ./
# Unduh semua dependensi
RUN go mod download

# Salin semua sisa kode sumber aplikasi kita
COPY . .

# Kompilasi aplikasi Go kita.
# CGO_ENABLED=0 membuat binary yang kompatibel dengan Alpine Linux yang minimalis.
# -o app akan memberi nama file hasil kompilasi menjadi 'app'.
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# --- Tahap 2: Final ---
# Kita gunakan image Alpine Linux yang sangat kecil sebagai kapsul akhir.
FROM alpine:latest

WORKDIR /app

# Salin HANYA file kredensial dan file template.
# Kita tidak butuh kode Go lagi di sini, hanya hasil jadinya.
COPY templates ./templates
COPY credentials.json ./credentials.json

# Salin file aplikasi yang sudah dikompilasi dari tahap 'builder'.
COPY --from=builder /app/app .

# Beritahu Docker bahwa kapsul ini akan membuka port 1412.
EXPOSE 1412

# Perintah yang akan dijalankan saat kapsul ini dinyalakan.
CMD [ "./app" ]