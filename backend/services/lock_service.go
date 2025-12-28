package services

import (
	"context"
	"fmt"
	"movie-ticket-backend/database"
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
	key := fmt.Sprintf("lock:screening:%s:seat:%s", screeningID, seatID)

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
	key := fmt.Sprintf("lock:screening:%s:seat:%s", screeningID, seatID)
	return s.RDB.Del(ctx, key).Err()
}

// IsSeatLocked เช็คสถานะ
func (s *LockService) IsSeatLocked(screeningID, seatID string) (bool, string) {
	ctx := context.Background()
	key := fmt.Sprintf("lock:screening:%s:seat:%s", screeningID, seatID)

	val, err := s.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ""
	}
	if err != nil {
		return false, ""
	}
	return true, val
}

func (s *LockService) GetLockedSeats(screeningID string) (map[string]string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("lock:screening:%s:seat:*", screeningID)

	keys, err := s.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	lockedSeats := make(map[string]string)
	if len(keys) == 0 {
		return lockedSeats, nil
	}

	// Fetch values (UserIDs) for all keys
	values, err := s.RDB.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("lock:screening:%s:seat:", screeningID)
	for i, key := range keys {
		if len(key) > len(prefix) {
			seatID := key[len(prefix):]
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
		// Format: lock:screening:SCR_ID:seat:SEAT_ID
		var scrID, seatID string
		_, _ = fmt.Sscanf(key, "lock:screening:%s:seat:%s", &scrID, &seatID)

		// Note: Sscanf might fail with %s if no separators.
		// Manual parse for safety:
		// lock:screening:s1:seat:A1
		var prefix = "lock:screening:"
		var midPart = ":seat:"

		if len(key) > len(prefix) && contains(key, midPart) {
			parts := split(key, ":")
			if len(parts) >= 5 {
				scrID = parts[2]
				seatID = parts[4]

				fmt.Printf("Key Expired! Screening: %s, Seat: %s. Broadcasting unlock...\n", scrID, seatID)

				// WS Broadcast UNLOCK
				WSHub.Broadcast <- SeatUpdateMessage{
					ScreeningID: scrID,
					SeatID:      seatID,
					Status:      "AVAILABLE",
				}
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
