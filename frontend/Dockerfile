# Stage 1: Build
FROM oven/bun:1-alpine AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

# .env dibaca saat build untuk NUXT_PUBLIC_* variables
COPY .env .env

RUN bun run build

# Stage 2: Runtime
FROM oven/bun:1-alpine

WORKDIR /app

# Nuxt standalone output tidak butuh node_modules
COPY --from=builder /app/.output ./.output

EXPOSE 8002

# Jalankan Nuxt SSR server
CMD ["node", ".output/server/index.mjs"]
