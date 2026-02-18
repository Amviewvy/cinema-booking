package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoReady bool

func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		fmt.Println("⚠️ MONGO_URI not set")
		return
	}

	for {
		fmt.Println("⏳ Trying to connect Mongo...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		cancel()

		if err == nil {
			ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
			err = client.Ping(ctxPing, nil)
			cancelPing()

			if err == nil {
				fmt.Println("✅ Mongo connected")
				MongoClient = client
				MongoReady = true
				return
			}
		}

		fmt.Println("❌ Mongo not ready, retry in 3s...")
		time.Sleep(3 * time.Second)
	}
}
