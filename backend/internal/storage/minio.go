package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
)

var Client *minio.Client
var Bucket string

// Init initializes MinIO client and ensures bucket exists
func Init(cfg config.MinIOConfig) error {
	var err error
	Client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("minio connect: %w", err)
	}

	Bucket = cfg.Bucket
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create bucket if not exists
	exists, err := Client.BucketExists(ctx, Bucket)
	if err != nil {
		return fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		err = Client.MakeBucket(ctx, Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("minio create bucket: %w", err)
		}
		logger.Info("MinIO bucket created", "bucket", Bucket)

		// Set bucket policy to public read (for video/image serving)
		policy := `{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::` + Bucket + `/*"]
			}]
		}`
		err = Client.SetBucketPolicy(ctx, Bucket, policy)
		if err != nil {
			logger.Warn("Failed to set bucket policy", "err", err)
		}
	}

	logger.Info("MinIO connected", "endpoint", cfg.Endpoint, "bucket", Bucket)
	return nil
}

// Upload uploads a file to MinIO and returns the public URL
func Upload(ctx context.Context, reader io.Reader, size int64, filename string, contentType string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("minio client not initialized")
	}

	// Generate unique object path: type/date/snowflake_id.ext
	ext := filepath.Ext(filename)
	category := detectCategory(ext)
	objectName := fmt.Sprintf("%s/%s/%d%s",
		category,
		time.Now().Format("2006/01/02"),
		snowflake.GenID(),
		ext,
	)

	_, err := Client.PutObject(ctx, Bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio put: %w", err)
	}

	// Return the public URL
	// In production: use CDN URL. Here we use MinIO's direct URL.
	url := fmt.Sprintf("http://%s/%s/%s", Client.EndpointURL().Host, Bucket, objectName)
	return url, nil
}

func detectCategory(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".mp4", ".mov", ".avi", ".webm", ".mkv":
		return "videos"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "images"
	default:
		return "files"
	}
}

// GetURL returns the public URL for a given object path
func GetURL(objectName string) string {
	if Client == nil {
		return ""
	}
	return fmt.Sprintf("http://%s/%s/%s", Client.EndpointURL().Host, Bucket, objectName)
}
