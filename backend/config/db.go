package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Connected to MongoDB")
	DB = client
	return DB
}

func GetCollection(collectionName string) *mongo.Collection {
	return DB.Database(os.Getenv("DB_NAME")).Collection(collectionName)
}
