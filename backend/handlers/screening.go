package handlers

import (
	"context"
	"fmt"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"movie-ticket-backend/services"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Screening Handler ---
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
	lockedSeatsMap, _ := lockService.GetLockedSeats(req.MovieID, req.StartTime)

	// Merge Status
	seatsCopy := make([]models.Seat, len(screening.Seats))
	copy(seatsCopy, screening.Seats)

	for i := range seatsCopy {
		if userID, ok := lockedSeatsMap[seatsCopy[i].ID]; ok {
			if seatsCopy[i].Status == models.SeatAvailable {
				seatsCopy[i].Status = "LOCKED"
				seatsCopy[i].LockedBy = userID
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
