# Используйте образ Go, поддерживающий ARM
FROM golang:1.21-alpine as builder

# Установите рабочий каталог в контейнере
WORKDIR /app

# Скопируйте файлы модуля Go и скачайте зависимости
COPY go.mod go.sum ./
RUN go mod download

# Скопируйте исходный код вашего приложения
COPY . .

# Скомпилируйте приложение для архитектуры ARM
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o myapp ./cmd/shortener

# Используйте образ alpine, поддерживающий ARM для финального контейнера
FROM alpine:latest
RUN apk --no-cache add ca-certificates && adduser -D -g '' shortener

# Скопируйте скомпилированное приложение из предыдущего шага
COPY --chown=shortener --from=builder /app/cmd/shortener/myapp .

# Откройте порт, который использует ваше приложение
EXPOSE 8080

# Запустите приложение
CMD ["./myapp"]
