package auth

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	firebaseAuth "firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var FirebaseClient *firebaseAuth.Client

func InitFirebase() {
	opt := option.WithCredentialsFile("firebase-service-account.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Firebase init error: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Firebase auth error: %v", err)
	}

	FirebaseClient = client
	log.Println("✅ Firebase Admin Initialized")
}
