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

// CreateVaultItem - only owner can create their own vault item
func CreateVaultItem(c *gin.Context) {
    uid := c.GetString("user_id")
    userID, _ := primitive.ObjectIDFromHex(uid)

    var item models.VaultItem
    if err := c.ShouldBindJSON(&item); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    item.ID = primitive.NewObjectID()
    item.UserID = userID // associate owner
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()
    _, err := config.GetCollection("vault").InsertOne(context.Background(), item)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, item)
}

// GetVaultItems - only owner can list their items
func GetVaultItems(c *gin.Context) {
    uid := c.GetString("user_id")
    userID, _ := primitive.ObjectIDFromHex(uid)

    cur, err := config.GetCollection("vault").Find(context.Background(), bson.M{"user_id": userID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cur.Close(context.Background())

    var items []models.VaultItem
    for cur.Next(context.Background()) {
        var item models.VaultItem
        cur.Decode(&item)
        items = append(items, item)
    }
    c.JSON(http.StatusOK, items)
}

// GetVaultItem - only owner can retrieve their item
func GetVaultItem(c *gin.Context) {
    uid := c.GetString("user_id")
    userID, _ := primitive.ObjectIDFromHex(uid)
    id, _ := primitive.ObjectIDFromHex(c.Param("id"))

    var item models.VaultItem
    err := config.GetCollection("vault").FindOne(context.Background(), bson.M{
        "_id": id,
        "user_id": userID,
    }).Decode(&item)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
        return
    }
    c.JSON(http.StatusOK, item)
}

// UpdateVaultItem - only owner can update their item
func UpdateVaultItem(c *gin.Context) {
    uid := c.GetString("user_id")
    userID, _ := primitive.ObjectIDFromHex(uid)
    id, _ := primitive.ObjectIDFromHex(c.Param("id"))

    var item models.VaultItem
    if err := c.ShouldBindJSON(&item); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    item.UpdatedAt = time.Now()
    // Only update if user owns the item
    res, err := config.GetCollection("vault").UpdateOne(context.Background(),
        bson.M{"_id": id, "user_id": userID},
        bson.M{"$set": item},
    )
    if err != nil || res.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found or not owned"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

// DeleteVaultItem - only owner can delete their item
func DeleteVaultItem(c *gin.Context) {
    uid := c.GetString("user_id")
    userID, _ := primitive.ObjectIDFromHex(uid)
    id, _ := primitive.ObjectIDFromHex(c.Param("id"))

    res, err := config.GetCollection("vault").DeleteOne(context.Background(), bson.M{"_id": id, "user_id": userID})
    if err != nil || res.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found or not owned"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}