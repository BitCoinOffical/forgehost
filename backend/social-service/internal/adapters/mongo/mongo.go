package mongodb

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoConfig struct {
	MongoHost string
	MongoPort string
}

func NewMongoConnect(cfg *MongoConfig, dataBase string, collection string) (*mongo.Collection, error) {
	u := url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(cfg.MongoHost, cfg.MongoPort),
	}
	connString := u.String()
	client, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("client.Ping: %w", err)
	}

	coll := client.Database(dataBase).Collection(collection)

	return coll, nil
}
