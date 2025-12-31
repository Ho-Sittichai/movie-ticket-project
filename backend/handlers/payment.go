package handlers

import (
	"fmt"
	"movie-ticket-backend/services"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Payment Handlers ---

// StartPayment Handler
func StartPayment(c *gin.Context) {
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

	lockService := services.NewLockService()

	// 1. Check if already paying
	if lockService.HasPaymentLock(userID) {
		c.JSON(409, gin.H{"error": "Payment already in progress. Please try again in 5 minutes."})
		return
	}

	// 2. Resolve Screening (Reusing helper from seat.go in same package)
	screeningID, err := resolveScreeningID(req.MovieID, req.StartTime)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// 3. Extend Seat Locks FIRST (Ensure validity)
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
		c.JSON(409, gin.H{"error": "Failed to extend locks (seats might have expired)"})
		return
	}

	// 4. Set Payment Lock
	lockDuration := 5 * time.Minute
	expireAt := time.Now().Add(lockDuration)
	err = lockService.SetPaymentLock(userID, services.PaymentLockDetails{
		UserID:      userID,
		MovieID:     req.MovieID,
		ScreeningID: screeningID,
		StartTime:   req.StartTime,
		SeatIDs:     req.SeatIDs,
	}, lockDuration)

	if err != nil {
		services.LogError("SYSTEM_ERROR", userID, err, map[string]interface{}{"context": "set_payment_lock"})
		c.JSON(500, gin.H{"error": "Failed to set payment lock"})
		return
	}

	c.JSON(200, gin.H{
		"message":        "Payment started",
		"extended_count": extendedCount,
		"expire_at":      expireAt,
	})
}

// CancelPayment Handler
func CancelPayment(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(string)

	lockService := services.NewLockService()

	// 1. Release Lock (Deleting the key prevents the 'expired' event, so no log will be written)
	lockService.ReleasePaymentLock(userID)

	c.JSON(200, gin.H{"message": "Payment processed"})
}
