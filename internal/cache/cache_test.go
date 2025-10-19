package cache

import (
	"order-service/internal/model"
	"testing"
	"time"
)

func TestCacheOperations(t *testing.T) {
	cache := New()

	order := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		DateCreated: time.Now(),
	}

	cache.Set(order)
	
	retrieved, exists := cache.Get("test123")
	if !exists {
		t.Error("Expected order to exist in cache")
	}
	if retrieved.OrderUID != "test123" {
		t.Errorf("Expected order UID 'test123', got '%s'", retrieved.OrderUID)
	}

	_, exists = cache.Get("nonexistent")
	if exists {
		t.Error("Expected non-existent order to not be found")
	}

	// Test Size
	if size := cache.Size(); size != 1 {
		t.Errorf("Expected cache size 1, got %d", size)
	}
}
