# Используем официальный образ Go как базовый
FROM golang:1.22.5-alpine AS build

# Устанавливаем рабочий каталог в контейнере
WORKDIR /app

# Копируем go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем все исходные файлы
COPY . .

# Сборка бинарного файла
RUN go build -o main .

# Используем меньший образ для выполнения
FROM alpine:latest

# Устанавливаем рабочий каталог
WORKDIR /app/

# Копируем бинарный файл из предыдущего этапа
COPY --from=build /app/main .

# Команда для запуска приложения
CMD ["./main"] 
