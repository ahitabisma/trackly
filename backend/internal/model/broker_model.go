package model

import "time"

type Broker struct {
	Symbol string    `bson:"symbol"`
	Name   string    `bson:"name"`
	Time   time.Time `bson:"time"`
}
