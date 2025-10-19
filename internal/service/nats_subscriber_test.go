package service

import (
	"order-service/internal/model"
	"testing"
)

func TestNatsSubscriber_Creation(t *testing.T) {
	mockRepo := &MockRepository{}
	mockCache := &Cache{Orders: make(map[string]interface{})}
	
	subscriber := NewNatsSubscriber(mockRepo, mockCache, "test-cluster", "test-client")
	
	if subscriber == nil {
		t.Error("NewNatsSubscriber should return non-nil value")
	}
	
	if subscriber.cluster != "test-cluster" {
		t.Errorf("Expected cluster 'test-cluster', got '%s'", subscriber.cluster)
	}
}

func TestCache_Storage(t *testing.T) {
	cache := &Cache{Orders: make(map[string]interface{})}
	order := &model.Order{OrderUID: "test123"}
	
	cache.Orders[order.OrderUID] = order
	
	stored, exists := cache.Orders["test123"]
	if !exists {
		t.Error("Order should exist in cache")
	}
	
	if stored.(*model.Order).OrderUID != "test123" {
		t.Error("Stored order should have correct OrderUID")
	}
}