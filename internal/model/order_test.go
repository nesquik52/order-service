package model

import (
	"testing"
)

func TestOrderValidation_ValidOrder(t *testing.T) {
	order := &Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		Delivery: Delivery{
			Name: "Test User",
		},
		Items: []Item{
			{Name: "Test Item"},
		},
	}

	if err := order.Validate(); err != nil {
		t.Errorf("Valid order should not fail validation: %v", err)
	}
}

func TestOrderValidation_InvalidOrder(t *testing.T) {
	tests := []struct {
		name  string
		order *Order
	}{
		{
			name:  "empty order",
			order: &Order{},
		},
		{
			name: "missing order_uid",
			order: &Order{
				TrackNumber: "TRACK123",
				Delivery: Delivery{Name: "Test"},
				Items:      []Item{{Name: "Test"}},
			},
		},
		{
			name: "missing delivery name", 
			order: &Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Delivery:    Delivery{},
				Items:       []Item{{Name: "Test"}},
			},
		},
		{
			name: "empty items",
			order: &Order{
				OrderUID:    "test123", 
				TrackNumber: "TRACK123",
				Delivery:    Delivery{Name: "Test"},
				Items:       []Item{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.order.Validate(); err == nil {
				t.Error("Expected validation error but got none")
			}
		})
	}
}

func TestOrderFromJSON(t *testing.T) {
	jsonData := `{
		"order_uid": "test123",
		"track_number": "TRACK123",
		"delivery": {"name": "Test User"},
		"items": [{"name": "Test Item"}]
	}`

	var order Order
	err := order.FromJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Failed to parse valid JSON: %v", err)
	}

	if order.OrderUID != "test123" {
		t.Errorf("Expected OrderUID 'test123', got '%s'", order.OrderUID)
	}
}