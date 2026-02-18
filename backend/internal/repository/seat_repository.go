package repository

import (
	"context"
	"fmt"

	"backend/internal/database"
	"backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

func GetSeats(showID string) ([]models.Seat, error) {

	if database.MongoClient == nil {
		return nil, fmt.Errorf("database not ready")
	}
	collection := database.MongoClient.Database("cinema").Collection("seats")

	filter := bson.M{"show_id": showID}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var seats []models.Seat
	if err = cursor.All(context.Background(), &seats); err != nil {
		return nil, err
	}

	return seats, nil
}
