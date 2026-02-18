package service

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"time"
)

func LogEvent(log models.AuditLog) {
	collection := database.MongoClient.Database("cinema").Collection("audit_logs")

	log.Timestamp = time.Now()

	_, err := collection.InsertOne(context.Background(), log)
	if err != nil {

		println("⚠️ Audit log insert failed:", err.Error())
	}
}
