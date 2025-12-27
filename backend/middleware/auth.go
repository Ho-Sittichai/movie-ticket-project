package middleware

import (
	"context"
	"fmt"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expected format: "Bearer real-jwt-{userID}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		if !strings.HasPrefix(token, "real-jwt-") {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		userIDHex := strings.TrimPrefix(token, "real-jwt-")
		objID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid User ID in token"})
			c.Abort()
			return
		}

		// Check User Role in DB
		collection := database.Mongo.Collection("users")
		var user models.User
		err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			c.JSON(401, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.Role != models.RoleAdmin {
			fmt.Printf("Access Denied: User %s (%s) tried to access Admin API\n", user.Name, user.Role)
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		// Pass User to context if needed
		c.Set("user", user)
		c.Next()
	}
}
