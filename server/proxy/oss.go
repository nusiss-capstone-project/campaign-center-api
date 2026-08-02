package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
)

// ErrOSSNotConfigured is returned when OSS credentials or config are missing.
var ErrOSSNotConfigured = errors.New("oss is not configured")

// ObjectStorage uploads objects and builds public URLs.
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	PublicURL(key string) string
}

type aliyunOSSStorage struct {
	bucket        *oss.Bucket
	publicBaseURL string
	endpoint      string
	bucketName    string
}

var (
	objectStorageOnce sync.Once
	objectStorageInst ObjectStorage
)

// GetObjectStorage returns the singleton Aliyun OSS client (or a not-configured stub).
func GetObjectStorage() ObjectStorage {
	objectStorageOnce.Do(func() {
		objectStorageInst = newObjectStorageFromConfig()
	})
	return objectStorageInst
}

func newObjectStorageFromConfig() ObjectStorage {
	cfg := config.Config
	if cfg == nil || cfg.OSSConfig == nil {
		return notConfiguredObjectStorage{}
	}
	oc := cfg.OSSConfig
	if strings.TrimSpace(oc.AccessKeyID) == "" || strings.TrimSpace(oc.AccessKeySecret) == "" {
		return notConfiguredObjectStorage{}
	}
	client, err := oss.New(oc.Endpoint, oc.AccessKeyID, oc.AccessKeySecret)
	if err != nil {
		return notConfiguredObjectStorage{}
	}
	bucket, err := client.Bucket(oc.Bucket)
	if err != nil {
		return notConfiguredObjectStorage{}
	}
	return &aliyunOSSStorage{
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(oc.PublicBaseURL), "/"),
		endpoint:      strings.TrimSpace(oc.Endpoint),
		bucketName:    strings.TrimSpace(oc.Bucket),
	}
}

type notConfiguredObjectStorage struct{}

func (notConfiguredObjectStorage) PutObject(context.Context, string, io.Reader, int64, string) error {
	return ErrOSSNotConfigured
}

func (notConfiguredObjectStorage) PublicURL(string) string { return "" }

func (s *aliyunOSSStorage) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	options := []oss.Option{}
	if strings.TrimSpace(contentType) != "" {
		options = append(options, oss.ContentType(contentType))
	}
	if size >= 0 {
		options = append(options, oss.ContentLength(size))
	}
	logger := log.WithContext(ctx)
	logger.Infow("oss_put_object_start", "bucket", s.bucketName, "key", key, "size", size, "content_type", contentType)
	if err := s.bucket.PutObject(key, reader, options...); err != nil {
		logger.Errorw("oss_put_object_failed", "bucket", s.bucketName, "key", key, "error", err)
		return err
	}
	logger.Infow("oss_put_object_success", "bucket", s.bucketName, "key", key, "url", s.PublicURL(key))
	return nil
}

func (s *aliyunOSSStorage) PublicURL(key string) string {
	key = strings.TrimLeft(key, "/")
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + key
	}
	return fmt.Sprintf("https://%s.%s/%s", s.bucketName, s.endpoint, key)
}
