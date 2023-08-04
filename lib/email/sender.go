package email

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gin-gonic/gin"
)

// todo: update addresses to be correct
const (
	sender  = "mail@confesi.com"
	charSet = "UTF-8"
	profile = "confesi"
)

var (
	svc                       *ses.SES
	errorLoadingTemplate      = errors.New("error loading email template")
	ErrorNoLinkGeneratedError = errors.New("no link generated")
)

type email struct {
	Destination ses.Destination
	Message     ses.Message
	Source      string
}

func init() {
	newSession, err := session.NewSession()
	if err != nil {
		panic(fmt.Sprintf("error initializing AWS SES session: %s", err))
	}
	svc = ses.New(newSession)
}

func New() *email {
	return &email{
		Source:  "Confesi" + " " + "<" + sender + ">",
		Message: ses.Message{},
	}
}

func (e *email) To(addresses []string, ccs []string) *email {
	// Convert the string slices to slices of pointers to strings.
	var ccAddresses []*string
	for _, cc := range ccs {
		email := cc // Create a new variable inside the loop scope
		ccAddresses = append(ccAddresses, &email)
	}

	var toAddresses []*string
	for _, address := range addresses {
		email := address // Create a new variable inside the loop scope
		toAddresses = append(toAddresses, &email)
	}

	e.Destination = ses.Destination{
		CcAddresses: ccAddresses,
		ToAddresses: toAddresses,
	}
	return e
}

func (e *email) Subject(subject string) *email {
	e.Message.Subject = &ses.Content{
		Charset: aws.String(charSet),
		Data:    aws.String(subject),
	}
	return e
}

func (e *email) LoadPasswordResetTemplate(link string) (*email, error) {
	// Get the absolute path of the running executable.
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errorLoadingTemplate
	}

	htmlTemplatePath := filepath.Join(currentDir, "templates", "password_reset_email_template.html")

	htmlContent, err := ioutil.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	// Replace the placeholder with the actual link in the HTML template.
	htmlBody := string(htmlContent)
	htmlBody = strings.Replace(htmlBody, "{{ link }}", link, -1)

	textTemplatePath := filepath.Join(currentDir, "templates", "password_reset_email_template.txt")

	textContent, err := ioutil.ReadFile(textTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	// Replace the placeholder with the actual link in the text template.
	textBody := string(textContent)
	textBody = strings.Replace(textBody, "{{ link }}", link, -1)

	e.Message.Body = &ses.Body{
		Html: &ses.Content{
			Charset: aws.String(charSet),
			Data:    aws.String(htmlBody),
		},
		Text: &ses.Content{
			Charset: aws.String(charSet),
			Data:    aws.String(textBody),
		},
	}

	return e, nil
}

func (e *email) LoadVerifyEmailTemplate(link string) (*email, error) {
	// Get the absolute path of the running executable.
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errorLoadingTemplate
	}

	htmlTemplatePath := filepath.Join(currentDir, "templates", "verify_email_email_template.html")
	textTemplatePath := filepath.Join(currentDir, "templates", "verify_email_email_template.txt")

	// Read the content of the HTML template file.
	htmlContent, err := ioutil.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	// Replace the placeholder with the actual link in the HTML template.
	htmlBody := string(htmlContent)
	htmlBody = strings.Replace(htmlBody, "{{ link }}", link, -1)

	// Read the content of the text template file.
	textContent, err := ioutil.ReadFile(textTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	// Replace the placeholder with the actual link in the text template.
	textBody := string(textContent)
	textBody = strings.Replace(textBody, "{{ link }}", link, -1)

	// Set the email body using the templates.
	e.Message.Body = &ses.Body{
		Html: &ses.Content{
			Charset: aws.String(charSet),
			Data:    aws.String(htmlBody),
		},
		Text: &ses.Content{
			Charset: aws.String(charSet),
			Data:    aws.String(textBody),
		},
	}

	return e, nil
}

func (e *email) Send() (*ses.SendEmailOutput, error) {
	return svc.SendEmail(&ses.SendEmailInput{
		Destination: &e.Destination,
		Message:     &e.Message,
		Source:      &e.Source,
	})
}

// Short-hand email sender

func SendVerificationEmail(c *gin.Context, authClient *auth.Client, userEmail string) error {

	link, err := authClient.EmailVerificationLink(c, userEmail)
	if err != nil {
		return ErrorNoLinkGeneratedError
	}
	em, err := New().
		To([]string{userEmail}, []string{}).
		Subject("Confesi Email Verification").
		LoadVerifyEmailTemplate(link)
	if err != nil {
		return err
	}
	_, err = em.Send()
	return err
}

func SendPasswordResetEmail(c *gin.Context, authClient *auth.Client, userEmail string) error {
	link, err := authClient.PasswordResetLink(c, userEmail)
	if err != nil {
		return ErrorNoLinkGeneratedError
	}
	em, err := New().
		To([]string{userEmail}, []string{}).
		Subject("Confesi Password Reset").
		LoadPasswordResetTemplate(link)
	if err != nil {
		return err
	}
	_, err = em.Send()
	return err
}
