package main

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func main() {

	opt := option.WithCredentialsFile("firebase-service-account.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		panic(err)
	}

	uid := "teBGyRbyAaYcQIxzr7pCFtuERjw2" // 👈 เอา UID มาใส่ตรงนี้

	err = client.SetCustomUserClaims(context.Background(), uid, map[string]interface{}{
		"role": "admin",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Admin role set successfully")
}
