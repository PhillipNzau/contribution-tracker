package config

import (
	"context"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoClient *mongo.Client
	DBName      string
	JWTSecret   []byte
	AESKey      []byte
}

func LoadConfig() (*Config, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "contribution"
	}
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		return nil, errors.New("JWT_SECRET required")
	}
	aes := os.Getenv("AES_KEY")
	if len(aes) != 32 {
		return nil, errors.New("AES_KEY must be exactly 32 bytes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	cfg := &Config{MongoClient: client, DBName: dbName, JWTSecret: []byte(jwt), AESKey: []byte(aes)}

	// ensure indexes
	// if err := ensureIndexes(cfg); err != nil {
	// 	log.Printf("index creation error: %v", err)
	// }

	return cfg, nil
}

// func ensureIndexes(cfg *Config) error {
// 	// db := cfg.MongoClient.Database(cfg.DBName)
// 	// users unique email
// 	// users := db.Collection("users")
// 	// _, err := users.Indexes().CreateOne(context.Background(), mongo.IndexModel{
// 	// 	Keys:    bsonD{{Key: "email", Value: 1}},
// 	// 	Options: options.Index().SetUnique(true),
// 	// })
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// credentials: user_id index
// 	// creds := db.Collection("credentials")
// 	// _, err = creds.Indexes().CreateOne(context.Background(), mongo.IndexModel{Keys: bsonD{{Key: "user_id", Value: 1}}})
// 	// if err != nil {
// 	// 	return err
// 	// }
	
// }

// small helper for building bson.D without importing bson repeatedly
type bsonD []struct{ Key string; Value interface{} }

func (b bsonD) MarshalBSON() ([]byte, error) { return nil, nil }
