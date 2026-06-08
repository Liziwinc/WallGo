# --- ЭТАП 1: Сборка бинарника ---
FROM golang:1.26-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь остальной код проекта
COPY . .

# Компилируем Go-приложение в один статически связанный бинарник

RUN CGO_ENABLED=0 GOOS=linux go build -o apiserver ./cmd/apiserver/main.go

# --- ЭТАП 2: Финальный легковесный образ ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из первого этапа
COPY --from=builder /app/apiserver .

# Копируем папку с фронтендом, чтобы бэкенд мог её раздавать
COPY --from=builder /app/web ./web

# Открываем порт, который слушает твой сервер
EXPOSE 8080

# Команда для запуска приложения при старте контейнера
CMD ["./apiserver"]