package fire

//! Package named `fire` because `firebase` is already taken many times by official packages.

import (
	"confesi/config"
	"context"
	"log"

	fb "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
	"firebase.google.com/go/v4/appcheck"
	"firebase.google.com/go/v4/internal"
	"google.golang.org/api/option"
)

var fbApp *FirebaseApp

type FirebaseApp struct {
	App        *fb.App
	AuthClient *auth.Client
	MsgClient  *messaging.Client
	AppCheck   *appcheck.Client
}

func init() {
	// Init Firebase app
	err := InitFirebase("firebase-secrets.json")
	if err != nil {
		// if we can't init firebase, we have an unrecoverable error
		log.Fatal("Error initializing Firebase app: ", err)
	}
}

func InitFirebase(secretsPath string) error {
	opt := option.WithCredentialsFile(secretsPath)
	app, err := fb.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return err
	}

	appCheck, err := appcheck.NewClient(context.Background(), &internal.AppCheckConfig{
		ProjectID: config.FirebaseProjectID,
	})
	if err != nil {
		return err
	}

	msgClient, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	fbApp = &FirebaseApp{
		App:        app,
		AuthClient: authClient,
		MsgClient:  msgClient,
		AppCheck:   appCheck,
	}
	return nil
}

func New() *FirebaseApp {
	return fbApp
}
