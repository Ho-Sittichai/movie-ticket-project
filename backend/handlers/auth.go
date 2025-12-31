package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"movie-ticket-backend/config"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Google OAuth
var googleOauthConfig *oauth2.Config

func getOAuthConfig() *oauth2.Config {
	if googleOauthConfig == nil {
		googleOauthConfig = &oauth2.Config{
			RedirectURL:  config.AppConfig.GoogleRedirectURL,
			ClientID:     config.AppConfig.GoogleClientID,
			ClientSecret: config.AppConfig.GoogleClientSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}
	}
	return googleOauthConfig
}

func GoogleLogin(c *gin.Context) {
	url := getOAuthConfig().AuthCodeURL("random-state") // Mock up code for call
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != "random-state" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	oauthConfig := getOAuthConfig()
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Code exchange failed"})
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User data fetch failed"})
		return
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON parse failed"})
		return
	}

	// Parse Google User Data
	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(userData, &googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON parse failed"})
		return
	}

	// Dump Data to Console (as requested)
	fmt.Printf("--- Google User Log ---\nID: %s\nEmail: %s\nName: %s\nPicture: %s\n-----------------------\n",
		googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Picture)

	// Upsert User in DB
	collection := database.Mongo.Collection("users")
	var user models.User

	// Check if user exists
	err = collection.FindOne(context.TODO(), bson.M{"email": googleUser.Email}).Decode(&user)

	if err != nil {
		// New User logic: Check if first user
		count, _ := collection.CountDocuments(context.TODO(), bson.M{})
		role := models.RoleUser
		if count == 0 {
			role = models.RoleAdmin // First user gets ADMIN
		}

		user = models.User{
			ID:         primitive.NewObjectID(),
			Email:      googleUser.Email,
			Name:       googleUser.Name,
			PictureURL: googleUser.Picture,
			Role:       role,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			c.JSON(500, gin.H{"error": "DB create error"})
			return
		}
	} else {
		// Update existing user info
		update := bson.M{
			"$set": bson.M{
				"name":        googleUser.Name,
				"picture_url": googleUser.Picture,
				"updated_at":  time.Now(),
			},
		}
		collection.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, update)

		// Refresh struct fields
		user.Name = googleUser.Name
		user.PictureURL = googleUser.Picture
	}

	// Generate Redirect (passing real data)
	appToken := "real-jwt-" + user.ID.Hex() // Use specific name to avoid collision

	redirectURL := fmt.Sprintf("http://localhost:5173/?google_auth=success&token=%s&user_id=%s&role=%s&name=%s&picture=%s&email=%s",
		appToken, user.ID.Hex(), user.Role, url.QueryEscape(user.Name), url.QueryEscape(user.PictureURL), url.QueryEscape(user.Email))

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
