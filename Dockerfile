# Используем официальный образ Go для сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем исходный код в контейнер
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o crypto_tracker ./cmd/main.go

RUN ls -la /app

# Используем минимальный образ Alpine для финального контейнера
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранный бинарный файл из builder
COPY --from=builder /app/crypto_tracker .

# Копируем файлы миграций
COPY migrations ./migrations

# Копируем файл .env (если используется)
COPY .env .

# Указываем порт, который будет использовать приложение
EXPOSE 8002

# Команда для запуска приложения
CMD ["./crypto_tracker"]