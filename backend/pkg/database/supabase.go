package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(host string, port int, database string, user string, password string) (*pgxpool.Pool, error) {
	encodedPassword := url.QueryEscape(password)

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		user,
		encodedPassword,
		host,
		port,
		database,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return db, nil
}
