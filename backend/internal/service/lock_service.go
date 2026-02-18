package service

import (
	"fmt"
	"time"

	"backend/internal/database"

	"github.com/redis/go-redis/v9"
)

func ValidateLockOwner(showID, seatID, userID string) (bool, error) {
	key := fmt.Sprintf("lock:show:%s:seat:%s", showID, seatID)

	val, err := database.RedisClient.Get(database.Ctx, key).Result()

	if err == redis.Nil {
		return false, nil // ไม่มี lock อยู่เลย
	}

	if err != nil {
		return false, err
	}

	return val == userID, nil
}

func ReleaseLock(showID, seatID string) error {
	key := fmt.Sprintf("lock:show:%s:seat:%s", showID, seatID)
	return database.RedisClient.Del(database.Ctx, key).Err()
}

func LockSeat(showID, seatID, userID string) (bool, error) {
	key := fmt.Sprintf("lock:show:%s:seat:%s", showID, seatID)

	result, err := database.RedisClient.SetNX(
		database.Ctx,
		key,
		userID,
		1*time.Minute,
		//20*time.Second,
	).Result()

	if err != nil {
		return false, err
	}

	return result, nil
}
