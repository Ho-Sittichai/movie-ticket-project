package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"movie-ticket-backend/config"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"movie-ticket-backend/services"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		// New User
		user = models.User{
			ID:         primitive.NewObjectID(),
			Email:      googleUser.Email,
			Name:       googleUser.Name,
			PictureURL: googleUser.Picture,
			Role:       models.RoleUser,
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

// --- Movie Handler ---
func GetMovies(c *gin.Context) {
	collection := database.Mongo.Collection("movies")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var movies []models.Movie
	if err = cursor.All(context.TODO(), &movies); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func CreateMovie(c *gin.Context) {
	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	movie.ID = primitive.NewObjectID()
	collection := database.Mongo.Collection("movies")
	_, err := collection.InsertOne(context.TODO(), movie)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, movie)
}

// --- Booking Handler ---
func GetScreeningDetails(c *gin.Context) {
	fmt.Println("GetScreeningDetails")
	var req struct {
		MovieID   string `json:"movie_id"`
		StartTime string `json:"start_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	movieObjID, err := primitive.ObjectIDFromHex(req.MovieID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid Movie ID"})
		return
	}

	// Parse incoming time
	reqTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		c.JSON(400, gin.H{"error": "Invalid Start Time format"})
		return
	}

	fmt.Printf("Searching for MovieID: %s, StartTime: %v\n", req.MovieID, reqTime)

	collection := database.Mongo.Collection("movies")
	var movie models.Movie
	err = collection.FindOne(context.TODO(), bson.M{"_id": movieObjID}).Decode(&movie)
	if err != nil {
		fmt.Println("Movie not found in DB")
		c.JSON(404, gin.H{"error": "Movie not found"})
		return
	}

	var screening *models.Screening
	for _, s := range movie.Screenings {
		fmt.Printf("Checking screening time: %v vs Req: %v\n", s.StartTime, reqTime)
		// Compare times (ignoring small differences if needed, but exact match preferred)
		if s.StartTime.Equal(reqTime) || s.StartTime.Format(time.RFC3339) == req.StartTime {
			screening = &s
			break
		}
	}

	if screening == nil {
		c.JSON(404, gin.H{"error": "Screening not found"})
		return
	}

	// Redis Lock check
	lockService := services.NewLockService()
	lockedSeats, _ := lockService.GetLockedSeats(screening.ID)
	lockedMap := make(map[string]bool)
	for _, sid := range lockedSeats {
		lockedMap[sid] = true
	}

	// Merge Status
	seatsCopy := make([]models.Seat, len(screening.Seats))
	copy(seatsCopy, screening.Seats)

	for i := range seatsCopy {
		if lockedMap[seatsCopy[i].ID] {
			if seatsCopy[i].Status == models.SeatAvailable {
				seatsCopy[i].Status = "LOCKED"
			}
		}
	}
	screening.Seats = seatsCopy

	c.JSON(200, gin.H{
		"screening": screening,
		"movie": gin.H{
			"id":           movie.ID,
			"title":        movie.Title,
			"duration_min": movie.DurationMin,
		},
	})
}

func LockSeat(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id"`
		MovieID   string `json:"movie_id"`
		StartTime string `json:"start_time"`
		SeatID    string `json:"seat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Resolve Screening ID from MovieID + StartTime
	screeningID, err := resolveScreeningID(req.MovieID, req.StartTime)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// 2. Lock Redis
	lockService := services.NewLockService()
	locked, err := lockService.LockSeat(screeningID, req.SeatID, req.UserID, 5*time.Minute)
	if err != nil {
		c.JSON(500, gin.H{"error": "Redis error"})
		return
	}
	if !locked {
		c.JSON(409, gin.H{"error": "Seat is currently selected"})
		return
	}

	// WS Broadcast
	services.WSHub.Broadcast <- services.SeatUpdateMessage{
		ScreeningID: screeningID, // Internal ID used for WS room/topic
		SeatID:      req.SeatID,
		Status:      "LOCKED",
	}

	c.JSON(200, gin.H{"message": "Seat locked"})
}

func BookSeat(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id"`
		MovieID   string `json:"movie_id"`
		StartTime string `json:"start_time"`
		SeatID    string `json:"seat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Resolve Screening ID
	screeningID, err := resolveScreeningID(req.MovieID, req.StartTime)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	lockService := services.NewLockService()
	locked, holder := lockService.IsSeatLocked(screeningID, req.SeatID)
	if !locked || holder != req.UserID {
		c.JSON(400, gin.H{"error": "Lock expired or invalid"})
		return
	}

	// Update Mongo
	// Update Screening.Seats[x].Status = "BOOKED"
	collection := database.Mongo.Collection("movies")

	// Use array filters to update nested element
	filter := bson.M{
		"screenings": bson.M{
			"$elemMatch": bson.M{
				"id":           screeningID,
				"seats.id":     req.SeatID,
				"seats.status": "AVAILABLE", // Concurrency check
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"screenings.$[scr].seats.$[seat].status": "BOOKED",
		},
	}

	arrayFilters := options.UpdateOptions{
		ArrayFilters: &options.ArrayFilters{
			Filters: []interface{}{
				bson.M{"scr.id": screeningID},
				bson.M{"seat.id": req.SeatID},
			},
		},
	}

	res, err := collection.UpdateOne(context.TODO(), filter, update, &arrayFilters)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(409, gin.H{"error": "Seat already booked or not found"})
		return
	}

	// Generate Booking Record
	booking := models.Booking{
		ID:          primitive.NewObjectID(),
		UserID:      req.UserID,
		ScreeningID: screeningID,
		SeatID:      req.SeatID,
		Status:      "SUCCESS",
		Amount:      120,
		CreatedAt:   time.Now(),
	}
	database.Mongo.Collection("bookings").InsertOne(context.TODO(), booking)

	lockService.UnlockSeat(screeningID, req.SeatID)

	services.GetQueueService().PublishEvent("BOOKING_SUCCESS", booking)

	c.JSON(200, gin.H{"message": "Booking Success", "booking_id": booking.ID})
}

// Helper to find internal Screening ID
func resolveScreeningID(movieIDHex, startTimeStr string) (string, error) {
	movieObjID, err := primitive.ObjectIDFromHex(movieIDHex)
	if err != nil {
		return "", fmt.Errorf("invalid Movie ID")
	}
	reqTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid Start Time")
	}

	collection := database.Mongo.Collection("movies")
	var movie models.Movie
	err = collection.FindOne(context.TODO(), bson.M{"_id": movieObjID}).Decode(&movie)
	if err != nil {
		return "", fmt.Errorf("movie not found")
	}

	for _, s := range movie.Screenings {
		if s.StartTime.Equal(reqTime) || s.StartTime.Format(time.RFC3339) == startTimeStr {
			return s.ID, nil // Return internal ID (e.g., s1)
		}
	}
	return "", fmt.Errorf("screening not found")
}
