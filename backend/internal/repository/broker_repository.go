package repository

import (
	"context"
	"fmt"
	"time"

	"trackly-backend/internal/model"
	"trackly-backend/pkg/filter"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BrokerRepository interface {
	FindAll(opt filter.Options, allowedFields []string) ([]model.Broker, int, error)
	Create(symbol string, name string) error
}

type brokerRepository struct {
	collection *mongo.Collection
}

func NewBrokerRepository(collection *mongo.Collection) BrokerRepository {
	return &brokerRepository{collection: collection}
}

func (r *brokerRepository) FindAll(opt filter.Options, allowedFields []string) ([]model.Broker, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Membuat filter MongoDB dari URL options
	mongoFilter, findOpts := filter.BuildMongoQuery(opt, allowedFields)

	// Hitung total data berdasarkan filter
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("count error: %w", err)
	}

	// Ambil data menggunakan filter dan option
	cur, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find error: %w", err)
	}
	defer cur.Close(ctx)

	var brokers []model.Broker
	for cur.Next(ctx) {
		var broker model.Broker
		if err := cur.Decode(&broker); err != nil {
			return nil, 0, err
		}
		brokers = append(brokers, broker)
	}
	return brokers, int(total), nil
}

func (r *brokerRepository) Create(symbol string, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	broker := model.Broker{Symbol: symbol, Name: name, Time: time.Now()}
	_, err := r.collection.InsertOne(ctx, broker)
	return err
}
