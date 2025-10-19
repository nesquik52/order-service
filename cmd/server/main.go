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

	connStr := "postgres://order_user:order_password@localhost:5432/orders?sslmode=disable"
	natsCluster := "test-cluster"
	natsClient := "order-service"
	natsChannel := "orders"

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	cache := cache.New()

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

	natsSubscriber := service.NewNatsSubscriber(repo, serviceCache, natsCluster, natsClient)
	if err := natsSubscriber.Subscribe(natsChannel); err != nil {
		log.Fatal("Failed to subscribe to NATS:", err)
	}
	log.Println("Subscribed to NATS channel:", natsChannel)

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		http.ServeFile(w, r, "web/templates/order.html")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"cache_size": cache.Size(),
		})
	})

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

	go func() {
		log.Println("Server starting on :8080")
		log.Println("Static files served from: web/static/")
		log.Println("Web interface: http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

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