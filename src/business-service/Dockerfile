FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git

WORKDIR /build

# Копируем go.mod и go.sum бизнес-сервиса
COPY business-service/go.mod business-service/go.sum ./

# Копируем go.mod auth-client в ожидаемое расположение
RUN mkdir -p /go/src/github.com/perpetua1g0d/bmstu-diploma
COPY auth-client /go/src/github.com/perpetua1g0d/bmstu-diploma/auth-client

# Обновляем replace-директиву для использования абсолютного пути
RUN go mod edit -replace github.com/perpetua1g0d/bmstu-diploma/auth-client=/go/src/github.com/perpetua1g0d/bmstu-diploma/auth-client

# Скачиваем зависимости
RUN go mod download

# Копируем весь код бизнес-сервиса
COPY business-service/ ./

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o business-service .

# Финальный образ
FROM alpine:latest
COPY --from=builder /build/business-service /business-service
CMD ["/business-service"]
