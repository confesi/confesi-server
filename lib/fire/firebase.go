package fire

//! Package named `fire` because `firebase` is already taken many times by official packages.

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var fb *FirebaseApp

type FirebaseApp struct {
	App        *firebase.App
	AuthClient *auth.Client
	MsgClient  *messaging.Client
}

func init() {
	// Init Firebase app
	err := InitFirebase("../../firebase-secrets.json")
	if err != nil {
		// if we can't init firebase, we have an unrecoverable error
		log.Fatal("Error initializing Firebase app: ", err)
	}
}

func InitFirebase(secretsPath string) error {
	opt := option.WithCredentialsFile(secretsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return err
	}

	msgClient, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	fb = &FirebaseApp{
		App:        app,
		AuthClient: authClient,
		MsgClient:  msgClient,
	}
	return nil
}

func New() *FirebaseApp {
	return fb
}
