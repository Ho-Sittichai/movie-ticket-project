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
	lockedSeatsMap, _ := lockService.GetLockedSeats(screening.ID)

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

func LockSeat(c *gin.Context) {
	// Get trusted UserID from Context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	UserID := userID.(string)

	var req struct {
		// UserID    string `json:"user_id"` // REMOVED: Use Trusted UserID
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

	// Check if already locked
	isLocked, holderID := lockService.IsSeatLocked(screeningID, req.SeatID)

	if isLocked {
		if holderID == UserID {
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
	locked, err := lockService.LockSeat(screeningID, req.SeatID, UserID, 5*time.Minute)
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
		UserID:      UserID,
		Status:      "LOCKED",
	}

	c.JSON(200, gin.H{"message": "Seat locked", "status": "LOCKED"})
}

func BookSeat(c *gin.Context) {
	// Get trusted UserID from Context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	UserID := userID.(string)

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
		if !locked || holder != UserID {
			fmt.Printf("Seat %s lock invalid for user %s\n", seatID, UserID)
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

		// Note: We use "AVAILABLE" in check, but if it's LOCKED by us, it might be visually "LOCKED" in DB?
		// Actually, in our logic, status in DB remains "AVAILABLE" until booked. Redis holds the "LOCKED" state.
		// Wait, did we update DB status to LOCKED?
		// In LockSeat: c.JSON(200, ... "status": "LOCKED"). Redis is set.
		// DB status is NOT updated to LOCKED in the current implementation (based on `GetScreeningDetails` merging).
		// So checking "seats.status": "AVAILABLE" is correct.

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
			UserID:      UserID,
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

	c.JSON(200, gin.H{"message": "Booking Success", "booked_count": bookedCount})
}

// ExtendSeatLock Handler for batch extension
func ExtendSeatLock(c *gin.Context) {
	// Get trusted UserID from Context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	trustedUserID := userID.(string)

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
		success, err := lockService.ExtendSeatLock(screeningID, seatID, trustedUserID, 5*time.Minute)
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
