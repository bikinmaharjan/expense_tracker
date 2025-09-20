FROM node:18 as frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM golang:1.21 as backend-builder

WORKDIR /app/backend
COPY backend/go.* ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=backend-builder /app/backend/main .
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080
CMD ["./main"]