# Order Service

Микросервис для обработки заказов с использованием NATS Streaming, PostgreSQL и in-memory кэша.

## Демонстрация работы

https://drive.google.com/file/d/1eDW5A7ShX4sErvBQeqaUmQlxfMxl3_Ch/view?usp=sharing

## Архитектура

- **NATS Streaming** - message broker для получения заказов
- **PostgreSQL** - основное хранилище данных  
- **In-memory cache** - кэш для быстрого доступа к данным
- **HTTP API** - REST интерфейс для получения заказов
- **Go** - язык реализации сервиса

# Быстрый старт

## Терминал:

**1 шаг. Запуск инфраструктуры (Docker)**

cd ~/order-service/order-service
docker-compose up -d

**2 шаг. Запуск сервера**

go run cmd/server/main.go

**3 шаг, отдельный терминал. Публикация тестового заказа**

go run cmd/publisher/main.go

**4 шаг. Проверка работы**

**JSON API**
curl "http://localhost:8080/order?id=b563feb7b2b84b6test"

**Web интерфейс (откройте в браузере)**
http://localhost:8080


**5 шаг. Тестирование**

go test ./... -v
