package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SeatStatus string

const (
	Available SeatStatus = "AVAILABLE"
	Locked    SeatStatus = "LOCKED"
	Booked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SeatID   string             `bson:"seat_id" json:"seat_id"`
	ShowID   string             `bson:"show_id" json:"show_id"`
	Status   SeatStatus         `bson:"status" json:"status"`
	LockedBy string             `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
}

// SeatLockRequest ใช้รับข้อมูลจาก Frontend ตอนกดเลือกที่นั่ง
type SeatLockRequest struct {
	ShowID string `json:"show_id" binding:"required"`
	SeatID string `json:"seat_id" binding:"required"`
}

// BookingConfirmRequest ใช้รับข้อมูลตอนยืนยันการชำระเงิน
type BookingConfirmRequest struct {
	ShowID string `json:"show_id" binding:"required"`
	SeatID string `json:"seat_id" binding:"required"`
}
