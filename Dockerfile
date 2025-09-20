FROM node:20 AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
COPY frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM golang:1.24 AS backend-builder

WORKDIR /app/backend
COPY backend/go.* ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=backend-builder /app/backend/main .
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080
CMD ["./main"]