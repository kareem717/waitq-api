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

type unitOfWork struct {
	accountRepo      *account.AccountRepository
	waitlistRepo     *waitlist.WaitlistRepository
	subscriptionRepo *subscription.SubscriptionRepository
	tx               *bun.Tx
}

func (u *unitOfWork) Account() storage.AccountRepository {
	return u.accountRepo
}

func (u *unitOfWork) Waitlist() storage.WaitlistRepository {
	return u.waitlistRepo
}

func (u *unitOfWork) Subscription() storage.SubscriptionRepository {
	return u.subscriptionRepo
}

func (u *unitOfWork) Commit() error {
	return u.tx.Commit()
}

func (u *unitOfWork) Rollback() error {
	return u.tx.Rollback()
}

type Repository struct {
	accountRepo      *account.AccountRepository
	waitlistRepo     *waitlist.WaitlistRepository
	subscriptionRepo *subscription.SubscriptionRepository
	db               *bun.DB
	ctx              context.Context
}

func NewRepository(config Config, ctx context.Context, logger *zap.Logger) *Repository {
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
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Attempting to ping the database...")
	err = db.PingContext(pingCtx)
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
	return &Repository{
		accountRepo:      account.NewAccountRepository(db, ctx),
		waitlistRepo:     waitlist.NewWaitlistRepository(db, ctx),
		subscriptionRepo: subscription.NewSubscriptionRepository(db, ctx),
		db:               db,
		ctx:              ctx,
	}
}

func (r *Repository) Account() storage.AccountRepository {
	return r.accountRepo
}

func (r *Repository) Waitlist() storage.WaitlistRepository {
	return r.waitlistRepo
}

func (r *Repository) Subscription() storage.SubscriptionRepository {
	return r.subscriptionRepo
}

func (r *Repository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) NewUnitOfWork() (storage.UnitOfWork, error) {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		return nil, err
	}

	return &unitOfWork{
		accountRepo:      account.NewAccountRepository(tx, r.ctx),
		waitlistRepo:     waitlist.NewWaitlistRepository(tx, r.ctx),
		subscriptionRepo: subscription.NewSubscriptionRepository(tx, r.ctx),
		tx:               &tx,
	}, nil
}

func (r *Repository) RunInTx(ctx context.Context, fn func(ctx context.Context, uow storage.UnitOfWork) error) error {
	uow, err := r.NewUnitOfWork()
	if err != nil {
		return err
	}

	err = fn(ctx, uow)
	if err != nil {
		uow.Rollback()
		return err
	}

	return uow.Commit()
}
