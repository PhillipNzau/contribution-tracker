package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phillip/contribution-tracker-go/config"
	"github.com/phillip/contribution-tracker-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListUsers(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by your auth middleware)
		// role, exists := c.Get("role")
		// if !exists || role != "admin" {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		// 	return
		// }

		col := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := col.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
			return
		}

		var users []models.User
		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode Users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetUser(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {

        usrID, err := primitive.ObjectIDFromHex(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
            return
        }

        var user models.User
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        err = cfg.MongoClient.Database(cfg.DBName).
            Collection("users").
            FindOne(ctx, bson.M{"_id": usrID}).
            Decode(&user)

        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found or not owned"})
            return
        }

        c.JSON(http.StatusOK, user)
    }
}


func UpdateUser(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var input struct {
			Name  string `json:"name,omitempty"`
			Email string `json:"email,omitempty"`
			Phone string `json:"phone,omitempty"`
			Role  string `json:"role,omitempty"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{}
		if input.Name != "" {
			update["name"] = input.Name
		}
		if input.Email != "" {
			update["email"] = input.Email
		}
		if input.Phone != "" {
			update["phone"] = input.Phone
		}
		if input.Role != "" {
			update["role"] = input.Role
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var updatedUser bson.M
		err = col.FindOneAndUpdate(
			ctx,
			bson.M{"_id": objID},
			bson.M{"$set": update},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&updatedUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "User updated successfully",
			"user":    updatedUser,
		})
	}
}



func DeleteUser(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role and userID from context
		role, _ := c.Get("role")
		requesterID, _ := c.Get("userID")

		// Get user id from URL param
		userID := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// If not admin, only allow deleting their own account
		if role != "admin" && requesterID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = col.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}
