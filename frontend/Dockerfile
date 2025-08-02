# Tahap 1: Build SvelteKit (dengan adapter-node)
FROM --platform=linux/arm64 node:18-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

# ⬇️ Tambahkan ini untuk menyalin file .env
COPY .env .env

COPY . .

RUN npm run build

# Tahap 2: Runtime container
FROM --platform=linux/arm64 node:18-alpine

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules

CMD ["node", "build"]
