package main

import (
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/model"
	"os"
	"time"

	"github.com/nats-io/stan.go"
)

type Config struct {
	NatsClusterID string
	NatsClientID  string
	NatsChannel   string
}

func loadConfig() *Config {
	return &Config{
		NatsClusterID: getEnv("NATS_CLUSTER_ID", "test-cluster"),
		NatsClientID:  getEnv("NATS_CLIENT_ID", "test-publisher"),
		NatsChannel:   getEnv("NATS_CHANNEL", "orders"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	cfg := loadConfig()

	log.Println("Starting publisher...")
	log.Printf("NATS Cluster: %s", cfg.NatsClusterID)
	log.Printf("NATS Channel: %s", cfg.NatsChannel)

	sc, err := stan.Connect(cfg.NatsClusterID, cfg.NatsClientID)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer sc.Close()

	log.Println("Connected to NATS successfully!")

	orders := generateTestOrders()

	for i, order := range orders {
		data, err := json.MarshalIndent(order, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal order %s: %v", order.OrderUID, err)
			continue
		}

		err = sc.Publish(cfg.NatsChannel, data)
		if err != nil {
			log.Printf("Failed to publish order %s: %v", order.OrderUID, err)
			continue
		}

		log.Printf("[%d/%d] Order %s published successfully!", i+1, len(orders), order.OrderUID)
		fmt.Printf("Published order: %s, Track: %s\n", order.OrderUID, order.TrackNumber)
		
		if i < len(orders)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Println("All orders published successfully!")
	fmt.Println("Publisher finished work. Check the server logs for incoming orders.")
}

func generateTestOrders() []model.Order {
	now := time.Now()
	
	return []model.Order{
		{
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: model.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: model.Payment{
				Transaction:  "b563feb7b2b84b6test",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDt:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []model.Item{
				{
					ChrtID:      9934930,
					TrackNumber: "WBILMTESTTRACK",
					Price:       453,
					Rid:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test",
			DeliveryService:   "meest",
			Shardkey:          "9",
			SmID:              99,
			DateCreated:       now,
			OofShard:          "1",
		},
		{
			OrderUID:    "a462fec8c3c95c7demo",
			TrackNumber: "RUEXP DEMO123",
			Entry:       "RUEXP",
			Delivery: model.Delivery{
				Name:    "Ivan Ivanov",
				Phone:   "+79161234567",
				Zip:     "101000",
				City:    "Moscow",
				Address: "Tverskaya st. 10",
				Region:  "Moscow",
				Email:   "ivanov@mail.ru",
			},
			Payment: model.Payment{
				Transaction:  "a462fec8c3c95c7demo",
				RequestID:    "req_12345",
				Currency:     "RUB",
				Provider:     "sberpay",
				Amount:       5420,
				PaymentDt:    1637911127,
				Bank:         "sber",
				DeliveryCost: 500,
				GoodsTotal:   4920,
				CustomFee:    0,
			},
			Items: []model.Item{
				{
					ChrtID:      8847531,
					TrackNumber: "RUEXP DEMO123",
					Price:       2460,
					Rid:         "cd5320198b875bf1demo",
					Name:        "Smartphone Case",
					Sale:        10,
					Size:        "M",
					TotalPrice:  2214,
					NmID:        5421897,
					Brand:       "CaseMaster",
					Status:      202,
				},
				{
					ChrtID:      8847532,
					TrackNumber: "RUEXP DEMO123",
					Price:       1500,
					Rid:         "ef6431209c986cg2demo",
					Name:        "Screen Protector",
					Sale:        20,
					Size:        "Universal",
					TotalPrice:  1200,
					NmID:        5421898,
					Brand:       "GlassPro",
					Status:      202,
				},
				{
					ChrtID:      8847533,
					TrackNumber: "RUEXP DEMO123",
					Price:       1200,
					Rid:         "gh7542310da097dh3demo",
					Name:        "USB-C Cable",
					Sale:        15,
					Size:        "1m",
					TotalPrice:  1020,
					NmID:        5421899,
					Brand:       "CableTech",
					Status:      202,
				},
			},
			Locale:            "ru",
			InternalSignature: "demo_signature",
			CustomerID:        "demo_user",
			DeliveryService:   "russian-post",
			Shardkey:          "5",
			SmID:              88,
			DateCreated:       now.Add(-1 * time.Hour),
			OofShard:          "0",
		},
		{
			OrderUID:    "c573ffd9d4da6d8sample",
			TrackNumber: "USPS SAMPLE456",
			Entry:       "USPS",
			Delivery: model.Delivery{
				Name:    "John Smith",
				Phone:   "+12025550123",
				Zip:     "10001",
				City:    "New York",
				Address: "5th Avenue 123",
				Region:  "NY",
				Email:   "john.smith@example.com",
			},
			Payment: model.Payment{
				Transaction:  "c573ffd9d4da6d8sample",
				RequestID:    "req_67890",
				Currency:     "USD",
				Provider:     "stripe",
				Amount:       8999,
				PaymentDt:    1637914527,
				Bank:         "chase",
				DeliveryCost: 0,
				GoodsTotal:   8999,
				CustomFee:    0,
			},
			Items: []model.Item{
				{
					ChrtID:      7756420,
					TrackNumber: "USPS SAMPLE456",
					Price:       8999,
					Rid:         "hi8653421eb1a8ei4sample",
					Name:        "Wireless Headphones",
					Sale:        0,
					Size:        "One Size",
					TotalPrice:  8999,
					NmID:        6654321,
					Brand:       "AudioPro",
					Status:      202,
				},
			},
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "john_sample",
			DeliveryService:   "usps",
			Shardkey:          "3",
			SmID:              77,
			DateCreated:       now.Add(-2 * time.Hour),
			OofShard:          "1",
		},
	}
}