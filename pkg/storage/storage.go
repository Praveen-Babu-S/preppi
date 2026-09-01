package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"preppi.com/pkg/config"
)

type Client struct {
	client *minio.Client
	bucket string
}

func New(cfg *config.Config) (*Client, error) {
	client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bucket := cfg.Storage.Bucket
	if bucket == "" {
		bucket = "preppi"
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make bucket: %w", err)
		}
	}

	return &Client{client: client, bucket: bucket}, nil
}

func (c *Client) Upload(ctx context.Context, reader io.Reader, size int64, objectName, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", objectName, err)
	}

	url, err := c.client.PresignedGetObject(ctx, c.bucket, objectName, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("presign object: %w", err)
	}

	return url.String(), nil
}

func (c *Client) UploadBytes(ctx context.Context, data []byte, objectName, contentType string) (string, error) {
	return c.Upload(ctx, bytes.NewReader(data), int64(len(data)), objectName, contentType)
}

func (c *Client) Delete(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete %s: %w", objectName, err)
	}
	return nil
}

func (c *Client) URL(objectName string) (string, error) {
	url, err := c.client.PresignedGetObject(context.Background(), c.bucket, objectName, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("presign %s: %w", objectName, err)
	}
	return url.String(), nil
}
