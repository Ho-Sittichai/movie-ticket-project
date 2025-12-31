package services

import (
	"context"
	"fmt"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingService struct{}

func NewBookingService() *BookingService {
	return &BookingService{}
}

// BookingResult holds the summary of the booking operation
type BookingResult struct {
	BookedCount        int
	SuccessfulBookings []models.Booking
}

// ProcessBooking handles the core logic of booking seats
func (s *BookingService) ProcessBooking(userID string, movieID string, screeningID string, startTime string, seatIDs []string, paymentID string) (*BookingResult, error) {
	lockService := NewLockService()
	collection := database.Mongo.Collection("movies")
	bookingCollection := database.Mongo.Collection("bookings")

	bookedCount := 0
	var successfulBookings []models.Booking

	// 1. Fetch the screening price from the database
	movieObjID, _ := primitive.ObjectIDFromHex(movieID)
	var movie models.Movie
	collection.FindOne(context.TODO(), bson.M{"_id": movieObjID}).Decode(&movie)

	var screeningPrice float64 = 0
	for _, s := range movie.Screenings {
		if s.ID == screeningID {
			screeningPrice = s.Price
			break
		}
	}
	// Fallback if price is 0 (should not happen with seeded data, but safety check)
	if screeningPrice == 0 {
		screeningPrice = 200
	}

	for _, seatID := range seatIDs {
		// 1. Check Lock
		locked, holder := lockService.IsSeatLocked(movieID, startTime, seatID)
		if !locked || holder != userID {
			fmt.Printf("Seat %s lock invalid for user %s\n", seatID, userID)
			continue
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
			fmt.Printf("Seat %s update failed (modified 0)\n", seatID)
			continue
		}

		// 3. Create Booking Record
		booking := models.Booking{
			ID:              primitive.NewObjectID(),
			UserID:          userID,
			ScreeningID:     screeningID,
			ScreenStartTime: startTime,
			SeatID:          seatID,
			Status:          "SUCCESS",
			PaymentID:       paymentID,
			Amount:          screeningPrice,
			CreatedAt:       time.Now(),
		}
		bookingCollection.InsertOne(context.TODO(), booking)

		bookedCount++
		successfulBookings = append(successfulBookings, booking)

		// 4. Unlock Redis
		lockService.UnlockSeat(movieID, startTime, seatID)

		// 5. Update WS
		WSHub.Broadcast <- SeatUpdateMessage{
			ScreeningID: screeningID,
			MovieID:     movieID,
			StartTime:   startTime,
			SeatID:      seatID,
			Status:      "BOOKED",
		}
	}

	if bookedCount == 0 {
		return nil, fmt.Errorf("failed to book any seats")
	}

	result := &BookingResult{
		BookedCount:        bookedCount,
		SuccessfulBookings: successfulBookings,
	}

	// 6. Post-booking actions (Email & Audit)
	// Group Email Event
	GetQueueService().PublishEvent("BOOKING_GROUP_SUCCESS", successfulBookings)

	// Audit Log
	LogInfo("BOOKING_SUCCESS", userID, map[string]interface{}{
		"movie_id":          movieID,
		"screen_id":         screeningID,
		"screen_start_time": startTime,
		"seat_ids":          seatIDs,
		"booked_count":      bookedCount,
		"payment_id":        paymentID,
	})

	// Release Payment Lock
	lockService.ReleasePaymentLock(userID)

	return result, nil
}
