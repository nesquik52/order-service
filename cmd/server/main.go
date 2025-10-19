package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"order-service/internal/cache"
	"order-service/internal/model"
	"order-service/internal/repository"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/stan.go"
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

	// Подключение к NATS Streaming
	sc, err := stan.Connect(natsCluster, natsClient)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer sc.Close()

	// Подписка на канал - используем ОДИН кэш
	_, err = sc.Subscribe(natsChannel, func(msg *stan.Msg) {
		var order model.Order
		
		// Валидация JSON
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			log.Printf("Invalid JSON received: %v", err)
			return
		}

		// Валидация данных заказа
		if err := order.Validate(); err != nil {
			log.Printf("Invalid order data: %v", err)
			return
		}

		// Сохранение в БД
		ctx := context.Background()
		if err := repo.CreateOrder(ctx, &order); err != nil {
			log.Printf("Failed to save order to DB: %v", err)
			return
		}

		// Сохранение в ОСНОВНОЙ кэш
		cache.Set(&order)
		
		log.Printf("Order %s processed successfully", order.OrderUID)
	}, stan.DurableName("order-service"))

	if err != nil {
		log.Fatal("Failed to subscribe to NATS:", err)
	}
	log.Println("Subscribed to NATS channel:", natsChannel)

	// Обработчик статических файлов (CSS, JS)
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// HTTP сервер
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			http.Error(w, `{"error": "Order ID is required"}`, http.StatusBadRequest)
			return
		}

		order, exists := cache.Get(orderID)
		if !exists {
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
		log.Printf("Order %s requested", orderID)
	})

	// Главная страница с красивым интерфейсом
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		// Отдаем HTML файл из templates
		http.ServeFile(w, r, "web/templates/order.html")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"cache_size": cache.Size(),
		})
	})

	// Метрики для мониторинга
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "ok",
			"cache_size":  cache.Size(),
			"timestamp":   time.Now().Unix(),
			"service":     "order-service",
		})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	// Запуск сервера в горутине
	go func() {
		log.Println("🚀 Server starting on :8080")
		log.Println("📁 Static files served from: web/static/")
		log.Println("🌐 Web interface: http://localhost:8080")
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