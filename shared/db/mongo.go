package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TripsCollection      = "trips"
	RideFaresCollection = "ride_fares"
)

type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoConfig() *MongoConfig {

	return &MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		Database: "uber-database",
	}

}

func NewMongoClient(ctx context.Context, cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("MONGODB_URI is not set")
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("MONGODB_DATABASE is not set")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")

	return client, nil
}

func GetDatabase(client *mongo.Client, cfg *MongoConfig) *mongo.Database {
	return client.Database(cfg.Database)
}
