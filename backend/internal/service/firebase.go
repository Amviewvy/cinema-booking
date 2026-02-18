package service

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func InitFirebase() error {
	ctx := context.Background()

	opt := option.WithCredentialsFile("firebase-service-account.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return err
	}

	FirebaseApp = app
	return nil
}

func VerifyFirebaseToken(idToken string) (string, error) {

	client, err := FirebaseApp.Auth(context.Background())
	if err != nil {
		return "", err
	}

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", err
	}

	return token.UID, nil
}
