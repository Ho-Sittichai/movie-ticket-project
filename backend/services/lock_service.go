package services

import (
	"context"
	"encoding/json"
	"fmt"
	"movie-ticket-backend/database"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type LockService struct {
	RDB *redis.Client
}

func NewLockService() *LockService {
	return &LockService{
		RDB: database.RDB,
	}
}

// LockSeat พยายาม Lock ที่นั่ง
func (s *LockService) LockSeat(screeningID, seatID, userID string, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:screening:%s:seat:%s", screeningID, seatID)

	// Value คือ UserID เพื่อบอกว่าใคร Lock
	success, err := s.RDB.SetNX(ctx, key, userID, duration).Result() // SetNX for Set if Not Exists
	if err != nil {
		return false, err
	}
	return success, nil // true = locked success
}

// UnlockSeat ปลด Lock
func (s *LockService) UnlockSeat(screeningID, seatID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:screening:%s:seat:%s", screeningID, seatID)
	return s.RDB.Del(ctx, key).Err()
}

// ExtendSeatLock ต่อเวลา
func (s *LockService) ExtendSeatLock(screeningID, seatID, userID string, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:screening:%s:seat:%s", screeningID, seatID)

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

// IsSeatLocked เช็คสถานะ
func (s *LockService) IsSeatLocked(screeningID, seatID string) (bool, string) {
	ctx := context.Background()
	key := fmt.Sprintf("seat_lock:screening:%s:seat:%s", screeningID, seatID)

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

	// 1. Set the expiring lock
	err = s.RDB.Set(ctx, key, val, duration).Err()
	if err != nil {
		return err
	}

	// 2. Set the persistent data (longer TTL 10m to be safe) for the listener
	return s.RDB.Set(ctx, dataKey, val, duration+5*time.Minute).Err()
}

func (s *LockService) ReleasePaymentLock(userID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("payment_lock:%s", userID)
	dataKey := fmt.Sprintf("payment_data:%s", userID)

	// Delete both
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
		return nil, nil // No lock
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

func (s *LockService) GetLockedSeats(screeningID string) (map[string]string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("seat_lock:screening:%s:seat:*", screeningID)

	keys, err := s.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	lockedSeats := make(map[string]string)
	if len(keys) == 0 {
		return lockedSeats, nil
	}

	// Fetch all values (UserIDs) associated with these keys
	values, err := s.RDB.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		// Key format: seat_lock:screening:<ScreeningID>:seat:<SeatID>
		// We know keys matched the pattern, so we can parse carefully or just split
		// Pattern: seat_lock:screening:%s:seat:*

		// Robust parsing:
		var prefix = fmt.Sprintf("seat_lock:screening:%s:seat:", screeningID)
		if len(key) > len(prefix) {
			seatID := key[len(prefix):]

			// value is interface{}, cast to string
			if val, ok := values[i].(string); ok {
				lockedSeats[seatID] = val
			}
		}
	}

	return lockedSeats, nil
}

// ListenForExpireRedis คอยฟัง Event ตอน Key หมดอายุ
func (s *LockService) ListenForExpireRedis() {
	ctx := context.Background()
	pubsub := s.RDB.Subscribe(ctx, "__keyevent@0__:expired")

	fmt.Println("Redis Expiration Listener started...")

	ch := pubsub.Channel()
	for msg := range ch {
		key := msg.Payload
		// Format: seat_lock:screening:SCR_ID:seat:SEAT_ID
		var scrID, seatID string
		_, _ = fmt.Sscanf(key, "seat_lock:screening:%s:seat:%s", &scrID, &seatID)

		// Note: Sscanf might fail with %s if no separators.
		// Manual parse for safety:
		// seat_lock:screening:s1:seat:A1
		var prefix = "seat_lock:screening:"
		var midPart = ":seat:"

		if len(key) > len(prefix) && contains(key, midPart) {
			parts := split(key, ":")
			if len(parts) >= 5 {
				scrID = parts[2]
				seatID = parts[4]

				fmt.Printf("Key Expired! Screening: %s, Seat: %s. Broadcasting unlock...\n", scrID, seatID)

				// [AUDIT LOG] Seat Auto Released (Expired)
				// Note: movie_id and start_time are not in the Redis key,
				// logging screening_id and seat_id as primary identifiers.
				LogInfo("SEAT_RELEASED", "SYSTEM", map[string]interface{}{
					"screen_id": scrID,
					"seat_id":   seatID,
					"reason":    "expired",
				})

				// WS Broadcast UNLOCK
				WSHub.Broadcast <- SeatUpdateMessage{
					ScreeningID: scrID,
					SeatID:      seatID,
					Status:      "AVAILABLE",
				}
			}
		} else if strings.HasPrefix(key, "payment_lock:") {
			// Format: payment_lock:USER_ID
			userID := strings.TrimPrefix(key, "payment_lock:")
			dataKey := fmt.Sprintf("payment_data:%s", userID)

			// Fetch details from shadow key
			val, err := s.RDB.Get(ctx, dataKey).Result()
			if err == nil {
				var details PaymentLockDetails
				if err := json.Unmarshal([]byte(val), &details); err == nil {
					// [AUDIT LOG] Booking Timeout
					fmt.Printf("Payment Lock Expired for User: %s. Logging timeout...\n", userID)
					LogInfo("BOOKING_TIMEOUT", userID, map[string]interface{}{
						"movie_id":          details.MovieID,
						"screen_id":         details.ScreeningID,
						"screen_start_time": details.StartTime,
						"seat_ids":          details.SeatIDs,
					})
				}
				// Cleanup shadow key
				s.RDB.Del(ctx, dataKey)
			}
		}
	}
}

// Simple helpers because we are in services package
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}
