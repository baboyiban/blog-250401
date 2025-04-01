FROM oven/bun:latest AS builder

WORKDIR /app

COPY package*.json ./

RUN bun install

COPY . .

RUN bun run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
