package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(host, dbname, user, password string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=require",
		host, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// func NewDatabase(host string, database string, user string, password string) (*pgxpool.Pool, error) {
// 	encodedPassword := url.QueryEscape(password)

// 	dsn := fmt.Sprintf(
// 		"postgresql://%s:%s@%s/%s?sslmode=require&channel_binding=require",
// 		user,
// 		encodedPassword,
// 		host,
// 		database,
// 	)

// 	config, err := pgxpool.ParseConfig(dsn)
// 	if err != nil {
// 		return nil, fmt.Errorf("gagal parse config: %v", err)
// 	}

// 	config.MaxConns = 10
// 	config.MinConns = 2
// 	config.MaxConnLifetime = time.Hour
// 	config.MaxConnIdleTime = 30 * time.Minute

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	db, err := pgxpool.NewWithConfig(ctx, config)
// 	if err != nil {
// 		return nil, fmt.Errorf("gagal membuat pool: %v", err)
// 	}

// 	return db, nil
// }
