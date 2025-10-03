package controllers

import (
	"context"
	"net/http"
	"time"

	config "github.com/phillip/contribution-tracker-go/config"
	models "github.com/phillip/contribution-tracker-go/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ListNotifications(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}

		filter := bson.M{"user_id": userID}
		if c.Query("unread") == "true" {
			filter["read"] = false
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("notifications")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch notifications"})
			return
		}

		var notifs []models.Notification
		if err := cursor.All(ctx, &notifs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decode error"})
			return
		}

		c.JSON(http.StatusOK, notifs)
	}
}

func MarkNotificationRead(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("notifications")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = col.UpdateOne(ctx,
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{"read": true}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
	}
}
