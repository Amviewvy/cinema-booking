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
			collection := database.MongoClient.Database("cinema").Collection("seats")

			cursor, _ := collection.Find(
				database.Ctx,
				bson.M{"status": models.Locked},
			)

			for cursor.Next(database.Ctx) {
				var seat models.Seat
				cursor.Decode(&seat)

				key := fmt.Sprintf("lock:show:%s:seat:%s", seat.ShowID, seat.SeatID)

				_, err := database.RedisClient.Get(database.Ctx, key).Result()

				if err == redis.Nil {
					// TTL หมดแล้ว → revert
					collection.UpdateOne(
						database.Ctx,
						bson.M{
							"seat_id": seat.SeatID,
							"show_id": seat.ShowID,
							"status":  models.Locked,
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
						"status":  "AVAILABLE",
					})
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()
}
