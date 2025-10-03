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
		// --- Authenticated user ---
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		// --- Bind form fields ---
		var input struct {
			Title        string   `form:"title" binding:"required"`
			Description  string   `form:"description"`
			Location     string   `form:"location"`
			TargetAmount float64  `form:"target_amount"`
			Deadline     *string  `form:"deadline"` // string for binding, convert later
		}

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// --- Parse deadline if provided ---
		var deadline *time.Time
		if input.Deadline != nil && *input.Deadline != "" {
			parsed, err := time.Parse(time.RFC3339, *input.Deadline)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format, use RFC3339"})
				return
			}
			deadline = &parsed
		}

		// --- Handle file uploads ---
		form, err := c.MultipartForm()
		if err != nil && err != http.ErrNotMultipart {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
			return
		}

		var imageURLs []string
		if form != nil {
			files := form.File["images"] // key must be "images"
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
					return
				}

				url, err := utils.UploadToCloudinary(file, fileHeader)
				file.Close()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "image upload failed",
						"details": err.Error(),
						"file":    fileHeader.Filename,
					})
					return
				}

				imageURLs = append(imageURLs, url)
			}
		}

		// --- Save event ---
		now := time.Now()
		event := models.Event{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			Title:        input.Title,
			Description:  input.Description,
			Location:     input.Location,
			TargetAmount: input.TargetAmount,
			Deadline:     deadline,
			Status:       "ACTIVE",
			Images:       imageURLs,
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

		c.JSON(http.StatusCreated, event)
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
		// ✅ Validate requester identity
		role := c.GetString("role")
		requesterID := c.GetString("user_id")
		if requesterID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// ✅ Validate Event ID
		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		// ✅ Fetch existing event
		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var existing models.Event
		if err := col.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}

		// ✅ Check permission
		if role != "admin" && existing.UserID.Hex() != requesterID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// ✅ Bind input (form-data for mixed text + file upload)
		var input struct {
			Title        string   `form:"title"`
			Description  string   `form:"description"`
			Location     string   `form:"location"`
			TargetAmount float64  `form:"target_amount"`
			Deadline     *string  `form:"deadline"`
			Status       string   `form:"status"`
			Images       []string `form:"images"` // existing image URLs to keep
		}

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ✅ Prepare update document
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
		if input.TargetAmount > 0 {
			update["target_amount"] = input.TargetAmount
		}
		if input.Status != "" {
			update["status"] = input.Status
		}
		if input.Deadline != nil && *input.Deadline != "" {
			parsed, err := time.Parse(time.RFC3339, *input.Deadline)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format, use RFC3339"})
				return
			}
			update["deadline"] = parsed
		}

		// ✅ Handle new image uploads (multipart form)
		newImageURLs := []string{}
		form, _ := c.MultipartForm()
		if form != nil {
			files := form.File["new_images"] // key = "new_images"
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open image"})
					return
				}
				url, err := utils.UploadToCloudinary(file, fileHeader)
				file.Close()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed", "details": err.Error()})
					return
				}
				newImageURLs = append(newImageURLs, url)
			}
		}

		// ✅ Merge images (keep provided + add new)
		if input.Images != nil || len(newImageURLs) > 0 {
			update["images"] = append(input.Images, newImageURLs...)
		}

		// ❗ Reject empty update
		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		// ✅ Apply update
		_, err = col.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update event"})
			return
		}

		// ✅ Fetch updated event
		var updated models.Event
		if err := col.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Event updated successfully",
			"event":   updated,
		})
	}
}


// ---------------- DELETE ----------------
func DeleteEvent(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Validate requester identity
		role := c.GetString("role")
		requesterID := c.GetString("user_id")
		if requesterID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// ✅ Validate event ID
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// ✅ Fetch existing event
		var existing models.Event
		if err := col.FindOne(ctx, bson.M{"_id": oid}).Decode(&existing); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}

		// ✅ Check permission
		if role != "admin" && existing.UserID.Hex() != requesterID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// ✅ Delete event
		res, err := col.DeleteOne(ctx, bson.M{"_id": oid})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}

		// 🔹 (Optional) TODO: Delete images from Cloudinary
		for _, img := range existing.Images {
			  utils.DeleteFromCloudinary(img)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "event deleted successfully",
			"id":      oid.Hex(),
		})
	}
}

