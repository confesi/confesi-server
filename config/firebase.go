package config

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type FirebaseApp struct {
	App        *firebase.App
	AuthClient *auth.Client
}

func InitFirebase(secretsPath string) (*FirebaseApp, error) {
	opt := option.WithCredentialsFile(secretsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}

	firebaseApp := &FirebaseApp{
		App:        app,
		AuthClient: authClient,
	}

	return firebaseApp, nil
}
