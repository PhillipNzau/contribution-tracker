package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	config "github.com/phillip/contribution-tracker-go/config"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get the Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
            return
        }
        // Remove "Bearer " prefix
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        claims := jwt.MapClaims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return cfg.JWTSecret, nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        userID, ok := claims["user_id"].(string)
        if !ok || userID == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
            return
        }
        // Set user_id in Gin context
        c.Set("user_id", userID)
        c.Next()
    }
}