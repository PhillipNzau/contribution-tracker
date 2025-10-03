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
func CreateContribution(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.Contribution
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validate event_id
		if input.EventID.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
			return
		}

		// check if event exists
		eventCol := cfg.MongoClient.Database(cfg.DBName).Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var event models.Event
		err := eventCol.FindOne(ctx, bson.M{"_id": input.EventID}).Decode(&event)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event not found"})
			return
		}

		// validate contribution amount
		if input.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
			return
		}

		now := time.Now()
		contribution := models.Contribution{
			ID:                primitive.NewObjectID(),
			EventID:           input.EventID,
			ContributorName:   input.ContributorName,
			ContributorContact: input.ContributorContact,
			Amount:            input.Amount,
			Currency:          input.Currency,
			Method:            input.Method,
			PaymentRef:        input.PaymentRef,
			Status:            "PENDING",
			ReceiptURL:        input.ReceiptURL,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("contributions")
		if _, err := col.InsertOne(ctx, contribution); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create contribution"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      contribution.ID.Hex(),
			"message": "contribution created",
		})
	}
}


// ---------------- LIST ----------------
func ListContributions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		col := cfg.MongoClient.Database(cfg.DBName).Collection("contributions")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// --- Build filter ---
		filter := bson.M{}
		if eventID := c.Query("event_id"); eventID != "" {
			if oid, err := primitive.ObjectIDFromHex(eventID); err == nil {
				filter["event_id"] = oid
			}
		}
		if status := c.Query("status"); status != "" {
			filter["status"] = status
		}

		// --- Fetch data ---
		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch contributions"})
			return
		}

		var contributions []models.Contribution
		if err := cursor.All(ctx, &contributions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode contributions"})
			return
		}

		if len(contributions) == 0 {
			c.JSON(http.StatusOK, []models.Contribution{})
			return
		}

		// --- Pick the most recently updated contribution ---
		latest := contributions[0]
		for _, ctn := range contributions {
			if ctn.UpdatedAt.After(latest.UpdatedAt) {
				latest = ctn
			}
		}

		// --- Generate ETag from latest contribution ---
		etag := utils.GenerateETag(latest.ID, latest.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		// --- Add Last-Modified from latest contribution ---
		c.Header("Last-Modified", latest.UpdatedAt.UTC().Format(http.TimeFormat))

		c.JSON(http.StatusOK, contributions)
	}
}


// ---------------- GET ----------------
func GetContribution(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contribution id"})
			return
		}

		var contribution models.Contribution
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = cfg.MongoClient.Database(cfg.DBName).
			Collection("contributions").
			FindOne(ctx, bson.M{"_id": oid}).
			Decode(&contribution)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "contribution not found"})
			return
		}

		etag := utils.GenerateETag(contribution.ID, contribution.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		c.JSON(http.StatusOK, contribution)
	}
}

// ---------------- UPDATE ----------------
func UpdateContribution(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contribution id"})
			return
		}

		var input models.Contribution
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"updated_at": time.Now()}
		if input.ContributorName != "" {
			update["contributor_name"] = input.ContributorName
		}
		if input.ContributorContact != "" {
			update["contributor_contact"] = input.ContributorContact
		}
		if input.Amount > 0 {
			update["amount"] = input.Amount
		}
		if input.Currency != "" {
			update["currency"] = input.Currency
		}
		if input.Method != "" {
			update["method"] = input.Method
		}
		if input.PaymentRef != "" {
			update["payment_reference"] = input.PaymentRef
		}
		if input.Status != "" {
			update["status"] = input.Status
		}
		if input.ReceiptURL != "" {
			update["receipt_url"] = input.ReceiptURL
		}

		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("contributions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update contribution"})
			return
		}
		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "contribution not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "contribution updated", "id": oid.Hex()})
	}
}

// ---------------- DELETE ----------------
func DeleteContribution(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contribution id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("contributions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := col.DeleteOne(ctx, bson.M{"_id": oid})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete contribution"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "contribution not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "contribution deleted", "id": oid.Hex()})
	}
}
