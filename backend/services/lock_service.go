package services

import (
	"context"
	"encoding/json"
	"fmt"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LockService struct {
	RDB *redis.Client
}

func NewLockService() *LockService {
	return &LockService{
		RDB: database.RDB,
	}
}

// LockSeat uses MovieID + StartTime + SeatID for unique locking
func (s *LockService) LockSeat(movieID, startTime, seatID, userID string, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:%s", movieID, startTime, seatID)

	// Value is UserID to indicate who holds the lock
	success, err := s.RDB.SetNX(ctx, key, userID, duration).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

// UnlockSeat uses MovieID + StartTime + SeatID
func (s *LockService) UnlockSeat(movieID, startTime, seatID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:%s", movieID, startTime, seatID)
	return s.RDB.Del(ctx, key).Err()
}

// ExtendSeatLock uses MovieID + StartTime + SeatID
func (s *LockService) ExtendSeatLock(movieID, startTime, seatID, userID string, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:%s", movieID, startTime, seatID)

	// Check ownership first
	val, err := s.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Not locked
	}
	if err != nil {
		return false, err
	}
	if val != userID {
		return false, nil // Locked by someone else
	}

	// Extend TTL
	return s.RDB.Expire(ctx, key, duration).Result()
}

// IsSeatLocked uses MovieID + StartTime + SeatID
func (s *LockService) IsSeatLocked(movieID, startTime, seatID string) (bool, string) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:%s", movieID, startTime, seatID)

	val, err := s.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ""
	}
	if err != nil {
		return false, ""
	}
	return true, val
}

// --- User Payment Lock ---

type PaymentLockDetails struct {
	UserID      string   `json:"user_id"`
	MovieID     string   `json:"movie_id"`
	ScreeningID string   `json:"screening_id"`
	StartTime   string   `json:"start_time"`
	SeatIDs     []string `json:"seat_ids"`
}

func (s *LockService) SetPaymentLock(userID string, details PaymentLockDetails, duration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("payment_lock:%s", userID)
	dataKey := fmt.Sprintf("payment_data:%s", userID)

	val, err := json.Marshal(details)
	if err != nil {
		return err
	}

	err = s.RDB.Set(ctx, key, val, duration).Err()
	if err != nil {
		return err
	}

	return s.RDB.Set(ctx, dataKey, val, duration+5*time.Minute).Err()
}

func (s *LockService) ReleasePaymentLock(userID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("payment_lock:%s", userID)
	dataKey := fmt.Sprintf("payment_data:%s", userID)

	s.RDB.Del(ctx, dataKey)
	return s.RDB.Del(ctx, key).Err()
}

func (s *LockService) HasPaymentLock(userID string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("payment_lock:%s", userID)
	count, _ := s.RDB.Exists(ctx, key).Result()
	return count > 0
}

func (s *LockService) GetPaymentLock(userID string) (*PaymentLockDetails, error) {
	ctx := context.Background()
	key := fmt.Sprintf("payment_lock:%s", userID)
	val, err := s.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var details PaymentLockDetails
	if err := json.Unmarshal([]byte(val), &details); err != nil {
		return nil, err
	}
	return &details, nil
}

// GetLockedSeats matches pattern for specific movie & time
func (s *LockService) GetLockedSeats(movieID, startTime string) (map[string]string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:*", movieID, startTime)

	keys, err := s.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	lockedSeats := make(map[string]string)
	if len(keys) == 0 {
		return lockedSeats, nil
	}

	values, err := s.RDB.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		// key: seat_lock:movie:<ID>:time:<Time>:seat:<SeatID>
		// Robust parsing tailored to this specific format
		prefix := fmt.Sprintf("seat_lock:movie:%s:time:%s:seat:", movieID, startTime)
		if strings.HasPrefix(key, prefix) {
			seatID := strings.TrimPrefix(key, prefix)
			if val, ok := values[i].(string); ok {
				lockedSeats[seatID] = val
			}
		}
	}

	return lockedSeats, nil
}

// ListenForExpireRedis Listens for key expiration
func (s *LockService) ListenForExpireRedis() {
	ctx := context.Background()
	pubsub := s.RDB.Subscribe(ctx, "__keyevent@0__:expired")

	fmt.Println("Redis Expiration Listener started...")

	ch := pubsub.Channel()
	for msg := range ch {
		key := msg.Payload

		// Key: seat_lock:movie:%s:time:%s:seat:%s
		if strings.HasPrefix(key, "seat_lock:movie:") {
			parts := strings.Split(key, ":")
			// seat_lock:movie:<ID>:time:<Time>:seat:<SeatID> -> 7 parts
			// 0: seat_lock, 1: movie, 2: <ID>, 3: time, 4: <Time>, 5: seat, 6: <SeatID>
			if len(parts) >= 7 {
				movieID := parts[2]
				// Time might contain colons (e.g. 2024-12-31T20:00:00Z)
				// So we take everything between 'time' and 'seat'
				// Use substring logic for safer parsing with variable separators
				// Pattern: seat_lock:movie:<MID>:time:<TIME>:seat:<SID>
				// Find first "time:" and last ":seat:"

				// Re-parsing strategy:
				// 1. Remove prefix "seat_lock:movie:"
				// 2. Find next ":time:"
				// 3. Find last ":seat:"

				remainder := strings.TrimPrefix(key, "seat_lock:movie:")
				timeSplit := strings.Index(remainder, ":time:")
				seatSplit := strings.LastIndex(remainder, ":seat:")

				if timeSplit != -1 && seatSplit != -1 && seatSplit > timeSplit {
					movieID = remainder[:timeSplit]
					startTime := remainder[timeSplit+6 : seatSplit]
					seatID := remainder[seatSplit+6:]

					// Resolve ScreeningID for WS Broadcast
					screeningID, err := s.getScreeningID(movieID, startTime)
					if err != nil {
						fmt.Printf("Failed to resolve screening ID for expired key %s: %v\n", key, err)
						continue
					}

					fmt.Printf("Key Expired! Movie: %s, Seat: %s. Broadcasting unlock...\n", movieID, seatID)

					LogInfo("SEAT_RELEASED", "SYSTEM", map[string]interface{}{
						"movie_id":  movieID,
						"seat_id":   seatID,
						"reason":    "expired",
						"screen_id": screeningID, // Log Internal ID too
					})

					WSHub.Broadcast <- SeatUpdateMessage{
						ScreeningID: screeningID,
						SeatID:      seatID,
						Status:      "AVAILABLE",
					}
				}
			}
		} else if strings.HasPrefix(key, "payment_lock:") {
			userID := strings.TrimPrefix(key, "payment_lock:")
			dataKey := fmt.Sprintf("payment_data:%s", userID)
			val, err := s.RDB.Get(ctx, dataKey).Result()
			if err == nil {
				var details PaymentLockDetails
				if err := json.Unmarshal([]byte(val), &details); err == nil {
					fmt.Printf("Payment Lock Expired for User: %s. Logging timeout...\n", userID)
					LogInfo("BOOKING_TIMEOUT", userID, map[string]interface{}{
						"movie_id":          details.MovieID,
						"screen_id":         details.ScreeningID,
						"screen_start_time": details.StartTime,
						"seat_ids":          details.SeatIDs,
					})
				}
				s.RDB.Del(ctx, dataKey)
			}
		}
	}
}

// Internal Helper to find Screening ID from MovieID + StartTime
func (s *LockService) getScreeningID(movieIDHex, startTimeStr string) (string, error) {
	movieObjID, err := primitive.ObjectIDFromHex(movieIDHex)
	if err != nil {
		return "", fmt.Errorf("invalid Movie ID")
	}

	// Try parsing standard time formats
	reqTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		// Try fallback if stored differently in key vs logic, but usually we stick to RFC3339
		return "", fmt.Errorf("invalid Start Time")
	}

	collection := database.Mongo.Collection("movies")
	var movie models.Movie
	err = collection.FindOne(context.TODO(), bson.M{"_id": movieObjID}).Decode(&movie)
	if err != nil {
		return "", fmt.Errorf("movie not found")
	}

	for _, sc := range movie.Screenings {
		// Compare time equality or string match
		if sc.StartTime.Equal(reqTime) || sc.StartTime.Format(time.RFC3339) == startTimeStr {
			return sc.ID, nil
		}
	}
	return "", fmt.Errorf("screening not found")
}
