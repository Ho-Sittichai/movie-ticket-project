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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- Seat Handlers ---
func LockSeat(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(string)

	var req struct {
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

	// Check for Payment Lock (Block changes if paying for THIS screening)
	paymentLock, _ := lockService.GetPaymentLock(userID)
	if paymentLock != nil {
		if paymentLock.MovieID == req.MovieID && paymentLock.StartTime == req.StartTime {
			c.JSON(409, gin.H{"error": "Cannot change seats while payment is in progress"})
			return
		}
	}

	// Check if already locked
	isLocked, holderID := lockService.IsSeatLocked(screeningID, req.SeatID)

	if isLocked {
		if holderID == userID {
			// Same user -> Unlock (Toggle)
			err := lockService.UnlockSeat(screeningID, req.SeatID)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to unlock"})
				return
			}

			// WS Broadcast UNLOCK
			services.WSHub.Broadcast <- services.SeatUpdateMessage{
				ScreeningID: screeningID, // Internal ID used for WS room/topic
				SeatID:      req.SeatID,
				Status:      "AVAILABLE",
			}

			c.JSON(200, gin.H{"message": "Seat unlocked", "status": "AVAILABLE"})
			return
		} else {
			// Different user -> Conflict
			c.JSON(409, gin.H{"error": "Seat is currently selected by another user"})
			return
		}
	}

	// Not locked -> Lock it
	locked, err := lockService.LockSeat(screeningID, req.SeatID, userID, 5*time.Minute)
	if err != nil {
		c.JSON(500, gin.H{"error": "Redis error"})
		return
	}
	if !locked {
		// Should have been caught by IsSeatLocked, but double check race condition
		c.JSON(409, gin.H{"error": "Seat is currently selected"})
		return
	}

	// WS Broadcast LOCK
	services.WSHub.Broadcast <- services.SeatUpdateMessage{
		ScreeningID: screeningID,
		SeatID:      req.SeatID,
		UserID:      userID,
		Status:      "LOCKED",
	}

	c.JSON(200, gin.H{"message": "Seat locked", "status": "LOCKED"})
}

func BookSeat(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(string)

	var req struct {
		MovieID   string   `json:"movie_id"`
		StartTime string   `json:"start_time"`
		SeatIDs   []string `json:"seat_ids"`
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
	collection := database.Mongo.Collection("movies")
	bookingCollection := database.Mongo.Collection("bookings")

	bookedCount := 0

	for _, seatID := range req.SeatIDs {
		// 1. Check Lock
		locked, holder := lockService.IsSeatLocked(screeningID, seatID)
		if !locked || holder != userID {
			fmt.Printf("Seat %s lock invalid for user %s\n", seatID, userID)
			continue // Skip this seat or abort? Skip for partial success preferrable here
		}

		// 2. Update Mongo (Set Status BOOKED)
		filter := bson.M{
			"screenings": bson.M{
				"$elemMatch": bson.M{
					"id":           screeningID,
					"seats.id":     seatID,
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
					bson.M{"seat.id": seatID},
				},
			},
		}

		res, err := collection.UpdateOne(context.TODO(), filter, update, &arrayFilters)
		if err != nil {
			fmt.Printf("Mongo Update Error for %s: %v\n", seatID, err)
			continue
		}
		if res.ModifiedCount == 0 {
			// This might happen if DB status turned BOOKED already
			fmt.Printf("Seat %s update failed (modified 0)\n", seatID)
			continue
		}

		// 3. Create Booking Record
		booking := models.Booking{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			ScreeningID: screeningID,
			SeatID:      seatID,
			Status:      "SUCCESS",
			Amount:      120, // Should fetch price from screening
			CreatedAt:   time.Now(),
		}
		bookingCollection.InsertOne(context.TODO(), booking)

		bookedCount++

		// 4. Unlock Redis
		lockService.UnlockSeat(screeningID, seatID)

		// 5. Events (Async to prevent blocking)
		go func(sID, stID string, b models.Booking) {
			services.GetQueueService().PublishEvent("BOOKING_SUCCESS", b)
			services.WSHub.Broadcast <- services.SeatUpdateMessage{
				ScreeningID: sID,
				SeatID:      stID,
				Status:      "BOOKED",
			}
		}(screeningID, seatID, booking)
	}

	if bookedCount == 0 {
		c.JSON(409, gin.H{"error": "Failed to book any seats (locks expired?)"})
		return
	}

	// Release Payment Lock
	lockService.ReleasePaymentLock(userID)

	c.JSON(200, gin.H{"message": "Booking Success", "booked_count": bookedCount})
}

// ExtendSeatLock Handler for batch extension
func ExtendSeatLock(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(string)

	var req struct {
		MovieID   string   `json:"movie_id"`
		StartTime string   `json:"start_time"`
		SeatIDs   []string `json:"seat_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	screeningID, err := resolveScreeningID(req.MovieID, req.StartTime)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	lockService := services.NewLockService()
	extendedCount := 0

	for _, seatID := range req.SeatIDs {
		success, err := lockService.ExtendSeatLock(screeningID, seatID, userID, 5*time.Minute)
		if err != nil {
			fmt.Printf("Error extending lock for seat %s: %v\n", seatID, err)
			continue
		}
		if success {
			extendedCount++
		}
	}

	if extendedCount == 0 && len(req.SeatIDs) > 0 {
		c.JSON(409, gin.H{"error": "Failed to extend locks (maybe expired?)"})
		return
	}

	c.JSON(200, gin.H{"message": "Locks extended", "count": extendedCount})
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
