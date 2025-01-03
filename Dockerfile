# Используем официальный образ Golang
FROM golang:1.23 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Статическая сборка
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main ./cmd/wallet/main.go

FROM scratch

WORKDIR /app

# Копируем собранный бинарник
COPY --from=builder /app/main .
COPY config ./config

# Открываем порт gRPC
EXPOSE 50051 50052 50053

# Команда по умолчанию
CMD ["./main", "--config-path=./config/local.yaml"]
