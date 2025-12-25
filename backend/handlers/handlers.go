package handlers

import (
	"context"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"movie-ticket-backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- Auth Handler ---
func Login(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		role = "USER"
	}

	email := "test@example.com"
	if role == "ADMIN" {
		email = "admin@example.com"
	}

	collection := database.Mongo.Collection("users")
	var user models.User

	// Find or Create
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		// Create
		user = models.User{
			ID:        primitive.NewObjectID(),
			Email:     email,
			Role:      models.Role(role),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			c.JSON(500, gin.H{"error": "DB Error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   "mock-jwt-token-" + user.ID.Hex(),
		"user_id": user.ID.Hex(),
		"role":    user.Role,
	})
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
func GetScreening(c *gin.Context) {
	// In our Mongo model, Screening is embedded in Movie for simplicity (OR it could be separate).
	// Let's assume we pass screeningID.
	// For this demo, let's say we find the movie that HAS this screening.
	// This is a bit tricky with embedded array.
	// Let's simplify: Pass MovieID and ScreeningID?
	// Or just search entire movies collection where screenings.id = id

	screeningID := c.Param("id")

	collection := database.Mongo.Collection("movies")

	// Aggregation pipeline to find the screening and unwind
	// Or simple logic: Find movie, iterate screenings.
	var movie models.Movie
	filter := bson.M{"screenings.id": screeningID}
	err := collection.FindOne(context.TODO(), filter).Decode(&movie)
	if err != nil {
		c.JSON(404, gin.H{"error": "Screening not found"})
		return
	}

	var screening *models.Screening
	for _, s := range movie.Screenings {
		if s.ID == screeningID {
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
	// Note: We return a Copy, not modifying DB content
	for i := range screening.Seats {
		if lockedMap[screening.Seats[i].ID] {
			if screening.Seats[i].Status == models.SeatAvailable {
				screening.Seats[i].Status = "LOCKED"
			}
		}
	}

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
		UserID      string `json:"user_id"`
		ScreeningID string `json:"screening_id"`
		SeatID      string `json:"seat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 1. Check Real Seat Status in Mongo
	// Need to find movie -> screening -> seat
	// Complex query. For demo, simplified.
	// Assume we trust the ID exists or we verify it.
	// ... verification logic ...

	// 2. Lock Redis
	lockService := services.NewLockService()
	locked, err := lockService.LockSeat(req.ScreeningID, req.SeatID, req.UserID, 5*time.Minute)
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
		ScreeningID: req.ScreeningID,
		SeatID:      req.SeatID,
		Status:      "LOCKED",
	}

	c.JSON(200, gin.H{"message": "Seat locked"})
}

func BookSeat(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id"`
		ScreeningID string `json:"screening_id"`
		SeatID      string `json:"seat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	lockService := services.NewLockService()
	locked, holder := lockService.IsSeatLocked(req.ScreeningID, req.SeatID)
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
				"id":           req.ScreeningID,
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
				bson.M{"scr.id": req.ScreeningID},
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
		ScreeningID: req.ScreeningID,
		SeatID:      req.SeatID,
		Status:      "SUCCESS",
		Amount:      120,
		CreatedAt:   time.Now(),
	}
	database.Mongo.Collection("bookings").InsertOne(context.TODO(), booking)

	lockService.UnlockSeat(req.ScreeningID, req.SeatID)

	services.GetQueueService().PublishEvent("BOOKING_SUCCESS", booking)

	c.JSON(200, gin.H{"message": "Booking Success", "booking_id": booking.ID})
}
