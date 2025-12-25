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
	success, err := s.RDB.SetNX(ctx, key, userID, duration).Result()
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

func (s *LockService) GetLockedSeats(screeningID string) ([]string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("lock:screening:%s:seat:*", screeningID)

	keys, err := s.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var lockedSeatIDs []string
	for _, key := range keys {
		// key: lock:screening:X:seat:Y
		// Extract Y. simpler to just split by :
		// lock screening X seat Y -> index 4
		var sID, seatID string
		fmt.Sscanf(key, "lock:screening:%s:seat:%s", &sID, &seatID)
		// Sscanf might be tricky with strings containing colons, simpler manual parse
		// But here IDs are likely simple.
		// Actually, let's use a simpler split since fmt.Sscanf with strings is weird if not space detached.

		// manual parse
		// assuming format is fixed
		// ... implementation detail ...
		// for demo, let's just assume we can get it.
		// Actually, let's just iterate and try to parse.
		// simplified:
		var prefix = fmt.Sprintf("lock:screening:%s:seat:", screeningID)
		if len(key) > len(prefix) {
			lockedSeatIDs = append(lockedSeatIDs, key[len(prefix):])
		}
	}

	return lockedSeatIDs, nil
}
