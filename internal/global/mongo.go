package global

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
)

func GetMongo() *mongo.Client {
	if mongoClient == nil {
		mongoOnce.Do(func() {
			uri := os.Getenv("MONGO_URI")
			if uri == "" {
				uri = "mongodb://localhost:27017"
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
			if err != nil {
				log.Fatalf("Failed to connect to MongoDB at %s: %v", uri, err)
			}

			if err := client.Ping(ctx, nil); err != nil {
				log.Fatalf("Failed to ping MongoDB at %s: %v. Ensure MongoDB is running.", uri, err)
			}

			mongoClient = client
		})
	}
	return mongoClient
}

func closeMongo() error {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return mongoClient.Disconnect(ctx)
	}
	return nil
}

func GetMongoDBName() string {
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "nfs_proxy"
	}
	return dbName
}
