package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"order-service/internal/cache"
	"order-service/internal/repository"
	"order-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Конфигурация
	connStr := "postgres://order_user:order_password@localhost:5432/orders?sslmode=disable"
	natsCluster := "test-cluster"
	natsClient := "order-service"
	natsChannel := "orders"

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализация кэша
	cache := cache.New()

	// Восстановление кэша из БД при запуске
	ctx := context.Background()
	orders, err := repo.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Warning: failed to restore cache from DB: %v", err)
	} else {
		cache.Restore(orders)
		log.Printf("Cache restored with %d orders", len(orders))
	}

	serviceCache := &service.Cache{
		Orders: make(map[string]interface{}),
	}

	// Инициализация NATS подписчика
	natsSubscriber := service.NewNatsSubscriber(repo, serviceCache, natsCluster, natsClient)
	if err := natsSubscriber.Subscribe(natsChannel); err != nil {
		log.Fatal("Failed to subscribe to NATS:", err)
	}
	log.Println("Subscribed to NATS channel:", natsChannel)

	// HTTP сервер
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		order, exists := cache.Get(orderID)
		if !exists {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
		log.Printf("Order %s requested", orderID)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Order Service</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.form { margin: 20px 0; }
				input { padding: 10px; width: 300px; }
				button { padding: 10px 20px; }
				.result { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
			</style>
		</head>
		<body>
			<h1>Order Service</h1>
			<div class="form">
				<input type="text" id="orderId" placeholder="Enter Order ID" value="b563feb7b2b84b6test">
				<button onclick="getOrder()">Get Order</button>
			</div>
			<div id="result" class="result"></div>
			
			<script>
				function getOrder() {
					const orderId = document.getElementById('orderId').value;
					fetch('/order?id=' + orderId)
						.then(response => {
							if (!response.ok) {
								throw new Error('Order not found');
							}
							return response.json();
						})
						.then(data => {
							document.getElementById('result').innerHTML = 
								'<h3>Order Data:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
						})
						.catch(error => {
							document.getElementById('result').innerHTML = 
								'<p style="color: red;">Error: ' + error + '</p>';
						});
				}
			</script>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"cache_size": cache.Size(),
		})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	// Запуск сервера в горутине
	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}