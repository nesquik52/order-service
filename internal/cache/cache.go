package cache

import (
	"order-service/internal/model"
	"sync"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func New() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

func (c *Cache) Set(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(orderUID string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[orderUID]
	return order, exists
}

func (c *Cache) GetAll() []*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	orders := make([]*model.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, order)
	}
	return orders
}

func (c *Cache) Restore(orders []*model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.orders = make(map[string]*model.Order)
	for _, order := range orders {
		c.orders[order.OrderUID] = order
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.orders)
}