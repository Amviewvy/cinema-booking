package main

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/websocket"
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func seedSeats() {
	if database.MongoClient == nil {
		fmt.Println("⚠️ Mongo not ready, skip seeding")
		return
	}

	collection := database.MongoClient.Database("cinema").Collection("seats")

	count, _ := collection.CountDocuments(context.Background(), bson.M{})
	if count > 0 {
		return
	}

	seats := []interface{}{}

	for row := 'A'; row <= 'C'; row++ {
		for num := 1; num <= 5; num++ {
			seat := models.Seat{
				SeatID: string(row) + string(rune('0'+num)),
				ShowID: "show1",
				Status: models.Available,
			}
			seats = append(seats, seat)
		}
	}

	if _, err := collection.InsertMany(context.Background(), seats); err != nil {
		fmt.Println("⚠️ Failed to seed seats:", err)
		return
	}
}

func DatabaseReadyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if database.MongoClient == nil {
			c.JSON(503, gin.H{"error": "database not ready"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	database.ConnectMongo()
	database.ConnectRedis()

	err := service.InitFirebase()
	if err != nil {
		fmt.Println("⚠️ Firebase initialization failed:", err)
	}
	seedSeats()

	service.StartLockCleanupWorker()
	websocket.StartBroadcast()
	auth.InitFirebase()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/seats", DatabaseReadyMiddleware(), func(c *gin.Context) {
		showID := c.DefaultQuery("show_id", "show1")

		seats, err := repository.GetSeats(showID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, seats)

	})

	r.GET("/admin/seats",
		middleware.FirebaseAuthMiddleware(),
		middleware.RequireRole("admin"),
		func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin access granted"})
		})

	admin := r.Group("/admin")
	admin.Use(middleware.FirebaseAuthMiddleware())
	admin.Use(middleware.RequireRole("admin"))
	admin.Use(DatabaseReadyMiddleware())

	admin.GET("/logs", func(c *gin.Context) {

		collection := database.MongoClient.
			Database("cinema").
			Collection("audit_logs")

		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var logs []models.AuditLog
		if err := cursor.All(context.Background(), &logs); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, logs)
	})

	r.GET("/admin/bookings",
		middleware.FirebaseAuthMiddleware(),
		middleware.RequireRole("admin"),
		func(c *gin.Context) {

			movie := c.Query("movie")
			date := c.Query("date")

			collection := database.MongoClient.
				Database("cinema").
				Collection("bookings")

			filter := bson.M{}

			if movie != "" {
				filter["movie"] = movie
			}

			if date != "" {
				filter["date"] = date
			}

			cursor, err := collection.Find(context.Background(), filter)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			var bookings []models.Booking
			if err := cursor.All(context.Background(), &bookings); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, bookings)
		})

	r.POST("/seats/lock",
		middleware.FirebaseAuthMiddleware(),
		func(c *gin.Context) {
			var req models.SeatLockRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "invalid request"})
				return
			}
			userID := c.MustGet("user_id").(string)

			if err := service.ProcessSeatLock(req, userID); err != nil {
				status := 500
				if err.Error() == "seat not available" || err.Error() == "seat already locked by someone else" {
					status = 409
				}

				service.LogEvent(models.AuditLog{
					Event:   "LOCK_FAILED",
					UserID:  userID,
					ShowID:  req.ShowID,
					SeatID:  req.SeatID,
					Message: err.Error(),
				})

				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			websocket.SendUpdate(gin.H{
				"event":   "seat_locked",
				"seat_id": req.SeatID,
				"status":  "LOCKED",
			})

			c.JSON(200, gin.H{"message": "seat locked"})
		})

	r.POST("/booking/confirm",
		middleware.FirebaseAuthMiddleware(),
		func(c *gin.Context) {
			var req models.BookingConfirmRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}

			userID := c.MustGet("user_id").(string)

			err := service.ConfirmBooking(req, userID)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			websocket.SendUpdate(gin.H{
				"event":   "seat_booked",
				"seat_id": req.SeatID,
				"status":  "BOOKED",
			})

			c.JSON(200, gin.H{"message": "การจองเสร็จสมบูรณ์!"})
		})

	r.POST("/payment/success",
		middleware.FirebaseAuthMiddleware(),
		func(c *gin.Context) {

			type Request struct {
				ShowID string `json:"show_id"`
				SeatID string `json:"seat_id"`
			}

			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "invalid request"})
				return
			}

			userID := c.GetString("user_id")

			// 1️⃣ ตรวจสอบ lock owner
			valid, err := service.ValidateLockOwner(req.ShowID, req.SeatID, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "lock validation failed"})
				return
			}
			if !valid {
				c.JSON(403, gin.H{"error": "not lock owner or lock expired"})
				return
			}

			collection := database.MongoClient.
				Database("cinema").
				Collection("seats")

			//  update LOCKED → BOOKED
			result, err := collection.UpdateOne(
				context.Background(),
				bson.M{
					"seat_id":   req.SeatID,
					"show_id":   req.ShowID,
					"status":    models.Locked,
					"locked_by": userID,
				},
				bson.M{
					"$set": bson.M{
						"status":    models.Booked,
						"locked_by": "",
					},
				},
			)

			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			if result.MatchedCount == 0 {
				c.JSON(400, gin.H{"error": "seat not in locked state"})
				return
			}

			bookingCollection := database.MongoClient.
				Database("cinema").
				Collection("bookings")

			_, err = bookingCollection.InsertOne(context.Background(), models.Booking{
				UserID: userID,
				ShowID: req.ShowID,
				SeatID: req.SeatID,
				Movie:  "Avengers", // หรือ map จาก show_id
				Date:   time.Now().Format("2006-01-02"),
				Status: "SUCCESS",
			})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			service.ReleaseLock(req.ShowID, req.SeatID)

			service.LogEvent(models.AuditLog{
				Event:   "BOOKING_SUCCESS",
				UserID:  userID,
				ShowID:  req.ShowID,
				SeatID:  req.SeatID,
				Message: "Seat successfully booked",
			})

			websocket.SendUpdate(gin.H{
				"event":   "seat_booked",
				"seat_id": req.SeatID,
				"status":  "BOOKED",
			})

			c.JSON(200, gin.H{
				"success": true,
				"message": "payment success",
			})
		},
	)

	r.GET("/ws", websocket.HandleWebSocket)

	r.Run(":8080")
}
