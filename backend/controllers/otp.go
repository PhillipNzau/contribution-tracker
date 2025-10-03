package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phillip/contribution-tracker-go/config"
	"github.com/phillip/contribution-tracker-go/models"
	"github.com/phillip/contribution-tracker-go/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func RequestOTP(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user models.User
		if err := users.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		expiry := time.Now().Add(10 * time.Minute)

		_, err := users.UpdateOne(ctx,
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"otp": otp, "otp_expiry": expiry}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save otp"})
			return
		}

		go utils.SendEmail(input.Email, "Your OTP Code", "Your OTP is: "+otp)

		c.JSON(http.StatusOK, gin.H{"message": "OTP sent to email"})
	}
}

// func VerifyOTP(cfg *config.Config) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var input struct {
// 			Email string `json:"email" binding:"required,email"`
// 			OTP   string `json:"otp" binding:"required"`
// 		}
// 		if err := c.ShouldBindJSON(&input); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		var user models.User
// 		if err := users.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user); err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid otp"})
// 			return
// 		}

// 		if user.OTP != input.OTP || time.Now().After(user.OTPExpiry) {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "otp expired or invalid"})
// 			return
// 		}

// 		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$unset": bson.M{"otp": "", "otp_expiry": ""}})

// 		accessToken, refreshToken, _ := createTokensForUser(user.ID, cfg)
// 		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"refresh_token": refreshToken}})

// 		c.JSON(http.StatusOK, gin.H{
// 			"access_token":  accessToken,
// 			"refresh_token": refreshToken,
// 			"user": gin.H{
// 				"id":    user.ID.Hex(),
// 				"name":  user.Name,
// 				"email": user.Email,
// 			},
// 		})
// 	}
// }
