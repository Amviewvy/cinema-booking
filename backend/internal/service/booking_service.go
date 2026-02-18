package service

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ConfirmBooking ย้าย Logic มาจาก main.go
func ConfirmBooking(req models.BookingConfirmRequest, userID string) error {
	// 1. ตรวจสอบว่า User คนนี้เป็นคนเดียวกับที่ Lock ที่นั่งไว้จริงไหม (ป้องกันคนอื่นมาสวมรอย)
	valid, err := ValidateLockOwner(req.ShowID, req.SeatID, userID)
	if err != nil || !valid {
		return errors.New("คุณไม่ใช่เจ้าของที่นั่งที่ล็อคไว้ หรือเวลาล็อคหมดอายุแล้ว")
	}

	seatsCollection := database.MongoClient.Database("cinema").Collection("seats")
	bookingCollection := database.MongoClient.Database("cinema").Collection("bookings")

	session, err := database.MongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	_, err = session.WithTransaction(context.Background(), func(sc mongo.SessionContext) (interface{}, error) {

		result, err := seatsCollection.UpdateOne(
			sc,
			bson.M{
				"seat_id":   req.SeatID,
				"show_id":   req.ShowID,
				"status":    models.Locked,
				"locked_by": userID,
			},
			bson.M{"$set": bson.M{
				"status":    models.Booked,
				"locked_by": userID,
			}},
		)

		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			return nil, errors.New("ไม่พบข้อมูลการล็อคที่นั่ง หรือสถานะไม่ถูกต้อง")
		}

		newBooking := models.Booking{
			UserID: userID,
			ShowID: req.ShowID,
			SeatID: req.SeatID,
			Movie:  "Avengers",
			Date:   time.Now().Format("2026-01-02"),
			Status: "SUCCESS",
		}

		_, err = bookingCollection.InsertOne(sc, newBooking)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	// 3. ลบ Lock ออกจาก Redis (ไม่ว่าจะแจ้งเตือนสำเร็จหรือไม่ก็ตาม)
	err = ReleaseLock(req.ShowID, req.SeatID)
	if err != nil {
		return err
	}

	if err != nil {
		LogEvent(models.AuditLog{
			Event:   "BOOKING_SUCCESS",
			UserID:  userID,
			ShowID:  req.ShowID,
			SeatID:  req.SeatID,
			Message: "Seat successfully booked",
		})
		return err
	}

	return nil
}
