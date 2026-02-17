package service

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
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

	// 2. อัปเดตสถานะที่นั่งใน MongoDB จาก LOCKED -> BOOKED
	result, err := seatsCollection.UpdateOne(
		context.Background(),
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
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("ไม่พบข้อมูลการล็อคที่นั่ง หรือสถานะไม่ถูกต้อง")
	}

	// 3. บันทึกข้อมูลการจอง (Booking History)
	newBooking := models.Booking{
		UserID: userID,
		ShowID: req.ShowID,
		SeatID: req.SeatID,
		Status: "SUCCESS",
	}
	_, err = bookingCollection.InsertOne(context.Background(), newBooking)
	if err != nil {
		return err
	}

	// 4. ลบ Key ใน Redis ออก เพราะจองสำเร็จแล้ว ไม่ต้องให้ Worker มา Cleanup อีก
	ReleaseLock(req.ShowID, req.SeatID)

	return nil
}
