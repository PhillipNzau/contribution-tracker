package config

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// EnsureCategoryIndexes creates indexes for the categories collection
// func EnsureCategoryIndexes(client *mongo.Client, dbName string) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	col := client.Database(dbName).Collection("categories")

// 	userIdx := mongo.IndexModel{
// 		Keys:    bson.D{{Key: "user_id", Value: 1}},
// 		Options: options.Index().SetBackground(true),
// 	}

// 	nameIdx := mongo.IndexModel{
// 		Keys:    bson.D{{Key: "name", Value: 1}},
// 		Options: options.Index().SetBackground(true),
// 	}

// 	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, nameIdx})
// 	if err != nil {
// 		log.Printf("⚠️ Could not create category indexes: %v", err)
// 	} else {
// 		log.Println("✅ Category indexes ensured")
// 	}
// }

// EnsureAllIndexes creates indexes for all collections
func EnsureAllIndexes(client *mongo.Client, dbName string) {
	// EnsureCategoryIndexes(client, dbName)
}
