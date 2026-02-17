package service

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessSeatLock(req models.SeatLockRequest, userID string) error {
	collection := database.MongoClient.Database("cinema").Collection("seats")

	var seat models.Seat
	err := collection.FindOne(context.Background(), bson.M{
		"seat_id": req.SeatID,
		"show_id": req.ShowID,
	}).Decode(&seat)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("seat not found")
		}
		return err
	}

	if seat.Status != models.Available {
		return errors.New("seat not available")
	}

	locked, err := LockSeat(req.ShowID, req.SeatID, userID)
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("seat already locked by someone else")
	}

	// อัปเดตสถานะใน MongoDB
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"seat_id": req.SeatID, "show_id": req.ShowID},
		bson.M{"$set": bson.M{"status": models.Locked, "locked_by": userID}},
	)
	return err
}
