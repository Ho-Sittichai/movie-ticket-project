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
	isLocked, holderID := lockService.IsSeatLocked(req.MovieID, req.StartTime, req.SeatID)

	if isLocked {
		if holderID == userID {
			// Same user -> Unlock (Toggle)
			err := lockService.UnlockSeat(req.MovieID, req.StartTime, req.SeatID)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to unlock"})
				return
			}

			// WS Broadcast UNLOCK
			services.WSHub.Broadcast <- services.SeatUpdateMessage{
				ScreeningID: screeningID, // Internal ID used for WS room/topic
				MovieID:     req.MovieID,
				StartTime:   req.StartTime,
				SeatID:      req.SeatID,
				Status:      "AVAILABLE",
			}

			// [AUDIT LOG] Seat Manually Released (Unlocked)
			services.LogInfo("SEAT_RELEASED", userID, map[string]interface{}{
				"movie_id":          req.MovieID,
				"screen_id":         screeningID,
				"screen_start_time": req.StartTime,
				"seat_id":           req.SeatID,
				"reason":            "user_unlocked",
			})

			c.JSON(200, gin.H{"message": "Seat unlocked", "status": "AVAILABLE"})
			return
		} else {
			// Different user -> Conflict
			c.JSON(409, gin.H{"error": "Seat is currently selected by another user"})
			return
		}
	}

	// Not locked -> Lock it
	locked, err := lockService.LockSeat(req.MovieID, req.StartTime, req.SeatID, userID, 5*time.Minute)
	if err != nil {
		services.LogError("SYSTEM_ERROR", userID, err, map[string]interface{}{"context": "redis_lock_seat"})
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
		ScreeningID: screeningID, // Internal ID used for WS room/topic
		MovieID:     req.MovieID,
		StartTime:   req.StartTime,
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
		PaymentID string   `json:"payment_id"` // [NEW] Payment Reference
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

	// Delegate to BookingService
	bookingService := services.NewBookingService()
	result, err := bookingService.ProcessBooking(userID, req.MovieID, screeningID, req.StartTime, req.SeatIDs, req.PaymentID)

	if err != nil {
		// Differentiate error types if needed, for now general 500 or 409
		// If "failed to book any seats" likely conflict
		if err.Error() == "failed to book any seats" {
			c.JSON(409, gin.H{"error": "Failed to book seats (already booked or lock missing)"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(200, gin.H{"message": "Booking Success", "booked_count": result.BookedCount})
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

	// screeningID resolution removed as it is not needed for ExtendSeatLock with new key format

	lockService := services.NewLockService()
	extendedCount := 0

	for _, seatID := range req.SeatIDs {
		success, err := lockService.ExtendSeatLock(req.MovieID, req.StartTime, seatID, userID, 5*time.Minute)
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
