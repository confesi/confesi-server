package utils

import (
	"log"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func VerifyEmail(client *auth.Client, ) {
	email := "user@example.com"
	link, err := client.EmailVerificationLinkWithSettings(, email, actionCodeSettings)
	if err != nil {
		log.Fatalf("error generating email link: %v\n", err)
	}
	return link
}

/*
email := "user@example.com"
link, err := client.EmailSignInLink(ctx, email, actionCodeSettings)
if err != nil {
        log.Fatalf("error generating email link: %v\n", err)
}

//! Construct sign-in with email link template, embed the link and send
//! using custom SMTP server.
sendCustomEmail(email, displayName, link)
*/
