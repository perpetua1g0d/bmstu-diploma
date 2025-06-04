FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git

WORKDIR /build
COPY . .

# Скачиваем зависимости
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o business-service .

# Финальный образ
FROM alpine:latest
WORKDIR /build
COPY --from=builder /build/business-service .

EXPOSE 8080
CMD ["./business-service"]
