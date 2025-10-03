package utils

import (
	"context"
	"time"

	config "github.com/phillip/contribution-tracker-go/config"
	models "github.com/phillip/contribution-tracker-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateNotification creates notifications for multiple recipients
func CreateNotification(cfg *config.Config, recipients []primitive.ObjectID, title, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notificationCol := cfg.MongoClient.Database(cfg.DBName).Collection("notifications")

	var docs []interface{}
	for _, r := range recipients {
		docs = append(docs, models.Notification{
			ID:        primitive.NewObjectID(),
			UserID:    r,
			Title:     title,
			Message:   message,
			Read:      false,
			CreatedAt: time.Now(),
		})
	}

	if len(docs) > 0 {
		_, err := notificationCol.InsertMany(ctx, docs)
		return err
	}
	return nil
}
