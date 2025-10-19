package service

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/model"
	"order-service/internal/repository"

	"github.com/nats-io/stan.go"
)

type NatsSubscriber struct {
	repo    repository.OrderRepository
	cache   *Cache
	cluster string
	client  string
}

type Cache struct {
	Orders map[string]interface{}
}

func NewNatsSubscriber(repo repository.OrderRepository, cache *Cache, cluster, client string) *NatsSubscriber {
	return &NatsSubscriber{
		repo:    repo,
		cache:   cache,
		cluster: cluster,
		client:  client,
	}
}

func (ns *NatsSubscriber) Subscribe(channel string) error {
	sc, err := stan.Connect(ns.cluster, ns.client)
	if err != nil {
		return err
	}

	_, err = sc.Subscribe(channel, func(msg *stan.Msg) {
		var order model.Order
		
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			log.Printf("Invalid JSON received: %v", err)
			return
		}

		if err := order.Validate(); err != nil {
			log.Printf("Invalid order data: %v", err)
			return
		}

		ctx := context.Background()
		if err := ns.repo.CreateOrder(ctx, &order); err != nil {
			log.Printf("Failed to save order to DB: %v", err)
			return
		}

		ns.cache.Orders[order.OrderUID] = &order
		
		log.Printf("Order %s processed successfully", order.OrderUID)
	}, stan.DurableName("order-service"))

	return err
}