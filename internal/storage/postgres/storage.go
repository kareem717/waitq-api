package postgres

import (
	"context"
	"errors"
	"log"
	"time"

	"waitq/api/internal/storage"
	"waitq/api/internal/storage/postgres/account"
	"waitq/api/internal/storage/postgres/subscription"
	"waitq/api/internal/storage/postgres/waitlist"

	"github.com/alexlast/bunzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/zap"
)

type Config struct {
	URL                   string
	MaxConnections        int32
	MinConnections        int32
	MaxConnectionIdleTime time.Duration
	MaxConnectionLifetime time.Duration
}

type ConfigOption func(*Config)

func NewConfig(url string, options ...ConfigOption) Config {
	config := Config{
		URL:                   url,
		MaxConnections:        10,
		MinConnections:        1,
		MaxConnectionIdleTime: 1 * time.Hour,
		MaxConnectionLifetime: 1 * time.Hour,
	}

	for _, option := range options {
		option(&config)
	}

	return config
}

func WithMaxConnections(maxConnections int32) ConfigOption {
	return func(c *Config) {
		c.MaxConnections = maxConnections
	}
}

func WithMinConnections(minConnections int32) ConfigOption {
	return func(c *Config) {
		c.MinConnections = minConnections
	}
}

func WithMaxConnectionIdleTime(maxConnectionIdleTime time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxConnectionIdleTime = maxConnectionIdleTime
	}
}

func WithMaxConnectionLifetime(maxConnectionLifetime time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxConnectionLifetime = maxConnectionLifetime
	}
}

func NewRepository(config Config, ctx context.Context, logger *zap.Logger) *storage.Repository {
	poolConfig, err := configDBPool(config)
	if err != nil {
		log.Fatalf("Error creating pool config: %v", err)
	}

	sqldb := stdlib.OpenDB(*poolConfig.ConnConfig)
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bunzap.NewQueryHook(bunzap.QueryHookOptions{
		Logger:       logger,
		SlowDuration: 200 * time.Millisecond, // Omit to log all operations as debug
	}))

	// Increase timeout duration
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Println("Attempting to ping the database...")
	err = db.PingContext(ctx)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			log.Fatalf("ping was canceled by the client: %v", err)
		case errors.Is(err, context.DeadlineExceeded):
			log.Fatalf("ping timed out: %v", err)
		default:
			log.Fatalf("ping failed: %v", err)
		}
	}

	log.Println("Successfully connected to the database.")
	return &storage.Repository{
		Account:      account.NewAccountRepository(db, ctx),
		Waitlist:     waitlist.NewWaitlistRepository(db, ctx),
		Subscription: subscription.NewSubscriptionRepository(db, ctx),
	}
}

func configDBPool(config Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnIdleTime = config.MaxConnectionIdleTime
	poolConfig.MaxConnLifetime = config.MaxConnectionLifetime

	return poolConfig, nil
}
