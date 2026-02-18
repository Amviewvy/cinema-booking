package service

import (
	"fmt"
	"time"

	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func StartLockCleanupWorker() {
	go func() {
		for {

			if database.MongoClient == nil {
				time.Sleep(5 * time.Second)
				continue
			}

			collection := database.MongoClient.
				Database("cinema").
				Collection("seats")

			cursor, err := collection.Find(
				database.Ctx,
				bson.M{"status": models.Locked},
			)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue

			}

			for cursor.Next(database.Ctx) {
				var seat models.Seat
				if err := cursor.Decode(&seat); err != nil {
					continue
				}

				key := fmt.Sprintf("lock:show:%s:seat:%s", seat.ShowID, seat.SeatID)

				_, err := database.RedisClient.Get(database.Ctx, key).Result()

				if err == redis.Nil {
					// TTL หมดแล้ว → revert
					fmt.Println("Seat released:", seat.SeatID)

					LogEvent(models.AuditLog{
						Event:   "BOOKING_TIMEOUT",
						UserID:  seat.LockedBy,
						ShowID:  seat.ShowID,
						SeatID:  seat.SeatID,
						Message: "Lock expired, seat auto released",
					})

					collection.UpdateOne(
						database.Ctx,
						bson.M{
							"seat_id":   seat.SeatID,
							"show_id":   seat.ShowID,
							"status":    models.Locked,
							"locked_by": seat.LockedBy,
						},
						bson.M{
							"$set": bson.M{
								"status":    models.Available,
								"locked_by": "",
							},
						},
					)

					websocket.SendUpdate(gin.H{
						"event":   "seat_released",
						"seat_id": seat.SeatID,
						"show_id": seat.ShowID,
						"status":  "AVAILABLE",
					})
				}
			}

			cursor.Close(database.Ctx)

			time.Sleep(5 * time.Second)
		}
	}()
}
