package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	config "github.com/phillip/contribution-tracker-go/config"
	models "github.com/phillip/contribution-tracker-go/models"
	utils "github.com/phillip/contribution-tracker-go/utils"
)

// =============================
// Register (send OTP only)
// =============================
func Register(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required,email"`
			Role  string `json:"role" binding:"required"`
			Phone  string `json:"phone" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check if email already exists
		count, _ := users.CountDocuments(ctx, bson.M{"email": input.Email})
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		// Check if phone already exists
		phone, _ := users.CountDocuments(ctx, bson.M{"phone": input.Phone})
		if phone > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "phone already registered"})
			return
		}

		user := models.User{
			ID:        primitive.NewObjectID(),
			Name:      input.Name,
			Email:     input.Email,
			Phone:     input.Phone,
			Role:     input.Role,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert new user
		if _, err := users.InsertOne(ctx, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}

		// Generate OTP
		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		expiry := time.Now().Add(10 * time.Minute)
		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"otp": otp, "otp_expiry": expiry}})

		// Send OTP
		body := utils.BuildOtpEmail(user.Email, otp)
		go utils.SendEmail(user.Email, "Verify your account", body)



		c.JSON(http.StatusCreated, gin.H{
			"status":  200,
			"message": "Registration successful, OTP sent to email",
			"user": gin.H{
				"id":    user.ID.Hex(),
				"name":  user.Name,
				"email": user.Email,
				"phone": user.Phone,
				"role": user.Role,
			},
		})
	}
}

// =============================
// Login (send OTP only)
// =============================
func Login(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email string `json:"email"` // can be email or phone
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user models.User
		filter := bson.M{}

		// Decide if input is email or phone
		if strings.Contains(input.Email, "@") {
			// Treat as email
			filter = bson.M{"email": input.Email}
		} else {
			// Treat as phone
			filter = bson.M{"phone": input.Email}
		}

		if err := users.FindOne(ctx, filter).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// Generate OTP
		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		expiry := time.Now().Add(10 * time.Minute)

		_, err := users.UpdateOne(ctx,
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"otp": otp, "otp_expiry": expiry}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save OTP"})
			return
		}

		// Send OTP (always to associated email)
		body := utils.BuildOtpEmail(user.Email, otp)
		go utils.SendEmail(user.Email, "Your Login OTP", body)

		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "OTP sent to email",
		})
	}
}


// =============================
// Verify OTP (issue tokens)
// =============================
func VerifyOTP(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email string `json:"email" binding:"required,email"`
			OTP   string `json:"otp" binding:"required"`
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid otp"})
			return
		}

		// Check OTP
		if user.OTP != input.OTP || time.Now().After(user.OTPExpiry) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "otp expired or invalid"})
			return
		}

		// Clear OTP
		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$unset": bson.M{"otp": "", "otp_expiry": ""}})

		// Create tokens
		accessToken, refreshToken, _ := createTokensForUser(user.ID, cfg)
		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"refresh_token": refreshToken}})

		c.JSON(http.StatusOK, gin.H{
			"status":        200,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"id":    user.ID.Hex(),
				"name":  user.Name,
				"email": user.Email,
				"phone": user.Phone,
				"role": user.Role,
			},
		})
	}
}

// =============================
// Refresh Token (unchanged)
// =============================
func RefreshToken(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh_token"})
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(input.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.JWTSecret, nil
		})
		if err != nil || !token.Valid || claims["type"] != "refresh" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		uid, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user models.User
		objID, _ := primitive.ObjectIDFromHex(uid)
		if err := users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if user.RefreshToken != input.RefreshToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token mismatch"})
			return
		}

		// Create new tokens
		accessToken, refreshToken, _ := createTokensForUser(user.ID, cfg)

		// Rotate refresh token
		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"refresh_token": refreshToken}})

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
}

// =============================
// Helpers
// =============================
func createTokensForUser(uid primitive.ObjectID, cfg *config.Config) (accessToken string, refreshToken string, err error) {
	// Access Token (short-lived)
	accessClaims := jwt.MapClaims{
		"user_id": uid.Hex(),
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = access.SignedString(cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (long-lived)
	refreshClaims := jwt.MapClaims{
		"user_id": uid.Hex(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refresh.SignedString(cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
