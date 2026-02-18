package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Booking struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID string             `bson:"user_id" json:"user_id"`
	ShowID string             `bson:"show_id" json:"show_id"`
	SeatID string             `bson:"seat_id" json:"seat_id"`
	Movie  string             `bson:"movie" json:"movie"`
	Date   string             `bson:"date" json:"date"`
	Status string             `bson:"status" json:"status"`
}
