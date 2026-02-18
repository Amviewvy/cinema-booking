package auth

import (
	"context"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func SetAdminRole(uid string) error {
	opt := option.WithCredentialsFile("firebase-service-account.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return err
	}

	claims := map[string]interface{}{
		"role": "admin",
	}

	return client.SetCustomUserClaims(context.Background(), uid, claims)
}
