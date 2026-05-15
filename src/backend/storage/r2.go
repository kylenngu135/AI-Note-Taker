package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var client *s3.Client

// testEndpoint overrides the R2 endpoint in tests. Set via SetTestEndpoint before calling InitR2.
var testEndpoint string

// SetTestEndpoint points the R2 client at a fake S3 server for tests.
func SetTestEndpoint(url string) {
	testEndpoint = url
}

func InitR2(accountID, accessKeyID, secretAccessKey string) error {
	endpoint := testEndpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			}, nil
		})),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if testEndpoint != "" {
			o.UsePathStyle = true
		}
	})
	return nil
}

func UploadTranscription(ctx context.Context, uploadID, text, bucket string) (string, error) {
	key := fmt.Sprintf("transcriptions/%s.txt", uploadID)

	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader([]byte(text)),
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload transcription: %w", err)
	}

	return key, nil
}

// UploadRawFile stores raw binary data at key "raw/{uploadID}" before processing.
func UploadRawFile(ctx context.Context, uploadID string, data []byte, contentType, bucket string) (string, error) {
	key := fmt.Sprintf("raw/%s", uploadID)
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload raw file: %w", err)
	}
	return key, nil
}

// DownloadFile retrieves an object from R2 and returns its contents as bytes.
func DownloadFile(ctx context.Context, key, bucket string) ([]byte, error) {
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file body: %w", err)
	}
	return data, nil
}

func DeleteTranscription(ctx context.Context, storageKey string) error {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("R2_BUCKET_NAME")),
		Key:    aws.String(storageKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete transcription: %w", err)
	}

	return nil
}
