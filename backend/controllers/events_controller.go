package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	config "github.com/phillip/contribution-tracker-go/config"
	models "github.com/phillip/contribution-tracker-go/models"
	utils "github.com/phillip/contribution-tracker-go/utils"
)

// ---------------- CREATE ----------------
func CreateEvent(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		var input models.Event
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		event := models.Event{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			Title:        input.Title,
			Description:  input.Description,
			Location:     input.Location,
			TargetAmount: input.TargetAmount,
			Deadline:     input.Deadline,
			Status:       "ACTIVE",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := col.InsertOne(ctx, event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create event"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": event.ID.Hex(), "message": "event created"})
	}
}

// ---------------- LIST ----------------
func ListEvents(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Validate user ID ---
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// --- Build filter ---
		filter := bson.M{"user_id": userID}
		if q := c.Query("q"); q != "" {
			filter["title"] = bson.M{"$regex": q, "$options": "i"}
		}

		// --- Fetch data ---
		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch events"})
			return
		}

		var events []models.Event
		if err := cursor.All(ctx, &events); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode events"})
			return
		}

		if len(events) == 0 {
			c.JSON(http.StatusOK, []models.Event{})
			return
		}

		// --- Pick the most recently updated event ---
		latest := events[0]
		for _, ev := range events {
			if ev.UpdatedAt.After(latest.UpdatedAt) {
				latest = ev
			}
		}

		// --- Generate ETag from latest event ---
		etag := utils.GenerateETag(latest.ID, latest.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		// --- Add Last-Modified from latest event ---
		c.Header("Last-Modified", latest.UpdatedAt.UTC().Format(http.TimeFormat))

		c.JSON(http.StatusOK, events)
	}
}

// ---------------- GET ----------------
func GetEvent(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		eventID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		var event models.Event
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = cfg.MongoClient.Database(cfg.DBName).
			Collection("events").
			FindOne(ctx, bson.M{"_id": eventID, "user_id": userID}).
			Decode(&event)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found or not owned"})
			return
		}

		etag := utils.GenerateETag(event.ID, event.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		c.JSON(http.StatusOK, event)
	}
}

// ---------------- UPDATE ----------------
func UpdateEvent(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		var input models.Event
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"updated_at": time.Now()}
		if input.Title != "" {
			update["title"] = input.Title
		}
		if input.Description != "" {
			update["description"] = input.Description
		}
		if input.Location != "" {
			update["location"] = input.Location
		}
		if input.TargetAmount != 0 {
			update["target_amount"] = input.TargetAmount
		}
		if input.Deadline != nil {
			update["deadline"] = input.Deadline
		}
		if input.Status != "" {
			update["status"] = input.Status
		}

		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := col.UpdateOne(ctx, bson.M{"_id": oid, "user_id": userID}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
			return
		}
		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "event updated", "id": oid.Hex()})
	}
}

// ---------------- DELETE ----------------
func DeleteEvent(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := col.DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "event deleted", "id": oid.Hex()})
	}
}
