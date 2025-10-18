package service

import (
	"context"
	"encoding/json"
	"order-service/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock репозитория
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) GetOrderByUID(ctx context.Context, orderUID string) (*model.Order, error) {
	args := m.Called(ctx, orderUID)
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockRepository) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Order), args.Error(1)
}

func TestOrderValidation(t *testing.T) {
	order := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		Delivery: model.Delivery{
			Name: "Test User",
		},
		Items: []model.Item{
			{
				Name: "Test Item",
			},
		},
	}

	err := order.Validate()
	assert.NoError(t, err)

	// Тест невалидного заказа
	invalidOrder := &model.Order{}
	err = invalidOrder.Validate()
	assert.Error(t, err)
}

func TestOrderJSON(t *testing.T) {
	order := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		DateCreated: time.Now(),
	}

	data, err := json.Marshal(order)
	assert.NoError(t, err)

	var decodedOrder model.Order
	err = json.Unmarshal(data, &decodedOrder)
	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, decodedOrder.OrderUID)
}