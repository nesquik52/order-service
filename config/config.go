package config

type Config struct {
    DatabaseURL   string
    NatsClusterID string
    NatsClientID  string
    NatsChannel   string
    ServerPort    string
}

func Load() *Config {
    return &Config{
        DatabaseURL:   "postgres://order_user:order_password@localhost:5432/orders?sslmode=disable",
        NatsClusterID: "test-cluster",
        NatsClientID:  "order-service",
        NatsChannel:   "orders",
        ServerPort:    ":8080",
    }
}