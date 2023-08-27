package uploads

import (
	"bytes"
	"confesi/config"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var (
	s3svc          *s3.S3
	rekognitionSvc *rekognition.Rekognition
)

func init() {
	newSession, err := session.NewSession()
	if err != nil {
		panic(fmt.Sprintf("error initializing AWS session: %s", err))
	}
	s3svc = s3.New(newSession)
	rekognitionSvc = rekognition.New(newSession)
}

func Upload(file io.Reader, filename string) (string, error) {
	uuidName := uuid.New().String() + filepath.Ext(filename)

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file bytes: %v", err)
	}

	// detect file type to be either jpg, jpeg or png
	// if the file type is not one of these, return an error
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" && fileType != "image/jpg" && fileType != "image/png" {
		return "", fmt.Errorf("invalid file type: %v", fileType)
	}

	// check the image using AWS Rekognition BEFORE uploading to S3
	input := &rekognition.DetectModerationLabelsInput{
		Image: &rekognition.Image{
			Bytes: fileBytes,
		},
	}

	result, err := rekognitionSvc.DetectModerationLabels(input)
	if err != nil {
		return "", fmt.Errorf("failed to analyze image: %v", err)
	}

	for _, label := range result.ModerationLabels {
		if *label.Confidence > config.AwsRekognitionConfidenceThreshold {
			if *label.Name == "Explicit Nudity" || *label.Name == "Nudity" ||
				*label.Name == "Graphic Male Nudity" || *label.Name == "Graphic Female Nudity" ||
				*label.Name == "Sexual Activity" || *label.Name == "Partial Nudity" || *label.Name == "Suggestive" {

				// if inappropriate content is detected, return an error
				return "", fmt.Errorf("inappropriate content detected in image: %v", *label.Name)
			}
		}
	}

	// if the image is appropriate, upload it to our S3 bucket
	_, err = s3svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.AwsUserUploadsBucket),
		Key:    aws.String(uuidName),
		Body:   bytes.NewReader(fileBytes), // use the bytes read earlier
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	return config.AwsUserUploadsBucketBaseUrl + uuidName, nil
}
