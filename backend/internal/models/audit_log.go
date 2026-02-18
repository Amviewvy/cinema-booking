package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty"`
	Event     string                 `bson:"event"`
	UserID    string                 `bson:"user_id,omitempty"`
	ShowID    string                 `bson:"show_id,omitempty"`
	SeatID    string                 `bson:"seat_id,omitempty"`
	Message   string                 `bson:"message,omitempty"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty"`
	Timestamp time.Time              `bson:"timestamp"`
}
