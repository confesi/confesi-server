package email

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// todo: update addresses to be correct
const (
	sender  = "matthew.rl.trent@gmail.com"
	charSet = "UTF-8"
	profile = "confesi"
)

var (
	svc                  *ses.SES
	errorLoadingTemplate = errors.New("error loading email template")
)

type email struct {
	Destination ses.Destination
	Message     ses.Message
	Source      string
}

func init() {
	newSession, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
	})
	if err != nil {
		panic(fmt.Sprintf("error initializing AWS SES session: %s", err))
	}
	svc = ses.New(newSession)
}

func New() *email {
	return &email{
		Source:  sender,
		Message: ses.Message{},
	}
}

func (e *email) To(addresses []*string, ccs []*string) *email {
	e.Destination = ses.Destination{
		CcAddresses: ccs,
		ToAddresses: addresses,
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
	htmlTemplatePath := filepath.Join("..", "..", "..", "templates", "password_reset_email_template.html")

	htmlContent, err := ioutil.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	htmlBody := string(htmlContent)

	htmlBody = strings.Replace(htmlBody, "{{ link }}", link, -1)

	textTemplatePath := filepath.Join("..", "..", "..", "templates", "password_reset_email_template.txt")

	textContent, err := ioutil.ReadFile(textTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

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
	htmlTemplatePath := filepath.Join("..", "..", "..", "templates", "verify_email_email_template.html")

	htmlContent, err := ioutil.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

	htmlBody := string(htmlContent)

	htmlBody = strings.Replace(htmlBody, "{{ link }}", link, -1)

	textTemplatePath := filepath.Join("..", "..", "..", "templates", "verify_email_email_template.txt")

	textContent, err := ioutil.ReadFile(textTemplatePath)
	if err != nil {
		return nil, errorLoadingTemplate
	}

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

func (e *email) Send() (*ses.SendEmailOutput, error) {
	return svc.SendEmail(&ses.SendEmailInput{
		Destination: &e.Destination,
		Message:     &e.Message,
		Source:      &e.Source,
	})
}

// ! different kinds of errors; worth noting for the future?

// if err != nil {
// 	if aerr, ok := err.(awserr.Error); ok {
// 		switch aerr.Code() {
// 		case ses.ErrCodeMessageRejected:
// 			fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
// 		case ses.ErrCodeMailFromDomainNotVerifiedException:
// 			fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
// 		case ses.ErrCodeConfigurationSetDoesNotExistException:
// 			fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
// 		default:
// 			fmt.Println(aerr.Error())
// 		}
// 	} else {

// 		fmt.Println(err.Error())
// 	}
