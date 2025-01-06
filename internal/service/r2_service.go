package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/maheshrc27/scheduling-api/configs"
)

type R2Service struct {
	config cfg.Config
}

func NewR2Service(cfg cfg.Config) *R2Service {
	return &R2Service{config: cfg}
}

func (r *R2Service) R2Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r.config.R2.AccessKey, r.config.R2.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		slog.Info(err.Error())
		log.Fatal(err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r.config.R2.AccountID))
	})
}

// Function to upload file to Cloudflare R2 Storage
func (r *R2Service) UploadToR2(ctx context.Context, key string, file []byte, filetype string) error {
	// Create a PutObjectInput with the specified bucket, key, file content, and content type
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.config.R2.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: aws.String(filetype), // Set the content type to image/jpeg (change as needed)
	}

	r2Client := r.R2Client()

	// Upload the file to Cloudflare R2 Storage
	_, err := r2Client.PutObject(ctx, input)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}
