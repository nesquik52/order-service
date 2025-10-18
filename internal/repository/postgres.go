package repository

import (
	"context"
	"database/sql"
	"order-service/internal/model"

	_ "github.com/lib/pq"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByUID(ctx context.Context, orderUID string) (*model.Order, error)
	GetAllOrders(ctx context.Context) ([]*model.Order, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Вставка основного заказа
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
		                   customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	// Вставка доставки
	_, err = tx.ExecContext(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	// Вставка платежа
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment (order_uid, transaction, request_id, currency, provider, 
		                   amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}

	// Вставка товаров
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, 
			                   sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetOrderByUID(ctx context.Context, orderUID string) (*model.Order, error) {
	var order model.Order
	
	// Получение основного заказа
	err := r.db.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature, 
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return nil, err
	}

	// Получение доставки
	err = r.db.QueryRowContext(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return nil, err
	}

	// Получение платежа
	err = r.db.QueryRowContext(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, 
		       bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	// Получение товаров
	rows, err := r.db.QueryContext(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil
}

func (r *PostgresRepository) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT order_uid FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, err
		}
		
		order, err := r.GetOrderByUID(ctx, orderUID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}
