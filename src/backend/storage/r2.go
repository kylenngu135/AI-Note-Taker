package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
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
		Body:        strings.NewReader(text),
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload transcription: %w", err)
	}

	return key, nil
}

// UploadRawFile streams a file to R2 at key "raw/{uploadID}" before processing.
func UploadRawFile(ctx context.Context, uploadID string, r io.Reader, size int64, contentType, bucket string) (string, error) {
	key := fmt.Sprintf("raw/%s", uploadID)
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
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
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file body: %w", err)
	}
	return data, nil
}

// DownloadFileToTemp streams an R2 object directly to a local temp file.
// The caller is responsible for deleting the returned path when done.
func DownloadFileToTemp(ctx context.Context, key, bucket string) (string, error) {
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	tmpFile, err := os.CreateTemp("", "r2-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile.Name(), nil
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
