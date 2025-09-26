# Сборка фронтенда (Svelte)
FROM node:20 AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Сборка бэкенда (Go)
FROM golang:1.24.2 as backend
WORKDIR /app/backend
COPY backend/go.* ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /app/frontend/dist ./public
RUN go build -o main .

# Финальный образ
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=backend /app/backend/main .
COPY --from=backend /app/backend/public ./public
ENV PORT=8080
EXPOSE 8080
CMD ["./main"]