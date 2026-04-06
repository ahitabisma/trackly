package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabase(host string, database string, user string, password string) (*pgxpool.Pool, error) {
	encodedPassword := url.QueryEscape(password)

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=require&channel_binding=require",
		user,
		encodedPassword,
		host,
		database,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal parse config: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %v", err)
	}

	return db, nil
}
