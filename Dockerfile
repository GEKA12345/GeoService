# Используем официальный образ Go как базовый
FROM golang:1.19-alpine as builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем исходники приложения в рабочую директорию
COPY ["go.mod", "go.sum", "./"]
# Скачиваем все зависимости
RUN go mod download

COPY /proxy ./proxy

WORKDIR /app/proxy
# Собираем приложение
RUN go build -o main

# Начинаем новую стадию сборки на основе минимального образа
FROM alpine:latest

# Добавляем исполняемый файл из первой стадии в корневую директорию контейнера
COPY --from=builder /app/proxy/main /main
COPY proxy/docs/swagger.json /docs/swagger.json

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["/main"]
