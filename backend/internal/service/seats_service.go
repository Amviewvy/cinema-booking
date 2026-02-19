package service

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func ProcessSeatLock(req models.SeatLockRequest, userID string) error {
	collection := database.MongoClient.Database("cinema").Collection("seats")

	var seat models.Seat
	err := collection.FindOne(context.Background(), bson.M{
		"seat_id": req.SeatID,
		"show_id": req.ShowID,
	}).Decode(&seat)

	if err != nil {
		return err
	}

	key := fmt.Sprintf("lock:show:%s:seat:%s", req.ShowID, req.SeatID)

	val, err := database.RedisClient.Get(context.Background(), key).Result()

	if err == nil {
		if val == userID {
			return nil
		}
		return errors.New("seat is currently locked by another user")
	}

	if err != redis.Nil {
		return err
	}

	locked, err := LockSeat(req.ShowID, req.SeatID, userID)
	if err != nil {
		return err
	}

	if !locked {
		return errors.New("seat already locked")
	}
	lockExpirre := time.Now().Add(1 * time.Minute)

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{
			"seat_id": req.SeatID,
			"show_id": req.ShowID,
		},
		bson.M{
			"$set": bson.M{
				"status":      models.Locked,
				"locked_by":   userID,
				"lock_expire": lockExpirre,
			},
		},
	)

	return err
}
