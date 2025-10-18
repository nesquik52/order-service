package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"order-service/internal/cache"
)

type Handler struct {
	cache *cache.Cache
	tmpl  *template.Template
}

func NewHandler(cache *cache.Cache) *Handler {
	tmpl := template.Must(template.ParseFiles("web/templates/order.html"))
	return &Handler{
		cache: cache,
		tmpl:  tmpl,
	}
}

func (h *Handler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	orderUID := r.URL.Query().Get("id")
	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, exists := h.cache.Get(orderUID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) ShowOrderPage(w http.ResponseWriter, r *http.Request) {
	orderUID := r.URL.Query().Get("id")
	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, exists := h.cache.Get(orderUID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.tmpl.Execute(w, order)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"cache_size": h.cache.Size(),
	})
}