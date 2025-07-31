# --- Tahap 1: Builder ---
# Gunakan Node.js versi LTS (Long-Term Support)
FROM node:20-alpine AS builder

WORKDIR /app

# Salin file package.json dan install dependensi
COPY package*.json ./
RUN npm install

# Salin sisa kode sumber aplikasi
COPY . .

# Jalankan perintah build untuk menghasilkan file statis
RUN npm run build

# --- Tahap 2: Server (menggunakan Nginx) ---
FROM nginx:stable-alpine

# Hapus konfigurasi default Nginx
RUN rm /etc/nginx/conf.d/default.conf

# Salin hasil build dari tahap 'builder' ke direktori web Nginx
COPY --from=builder /app/build /usr/share/nginx/html

# Port yang akan diekspos oleh Nginx
EXPOSE 80

# Perintah untuk menjalankan Nginx saat kontainer dimulai
CMD ["nginx", "-g", "daemon off;"]