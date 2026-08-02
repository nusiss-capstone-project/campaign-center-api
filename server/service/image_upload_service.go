package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
)

const (
	maxImageUploadBytes = 5 << 20 // 5MB
)

var allowedImageContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// ImageUploadService uploads images to object storage.
type ImageUploadService interface {
	Upload(ctx context.Context, filename string, reader io.Reader, size int64) (*data.ImageUploadData, error)
}

type imageUploadService struct {
	storage proxy.ObjectStorage
}

var (
	imageUploadServiceOnce sync.Once
	imageUploadServiceInst ImageUploadService
)

// NewImageUploadService builds an upload service (for tests).
func NewImageUploadService(storage proxy.ObjectStorage) ImageUploadService {
	return &imageUploadService{storage: storage}
}

// GetImageUploadService returns the singleton image upload service.
func GetImageUploadService() ImageUploadService {
	imageUploadServiceOnce.Do(func() {
		imageUploadServiceInst = NewImageUploadService(proxy.GetObjectStorage())
	})
	return imageUploadServiceInst
}

func (s *imageUploadService) Upload(ctx context.Context, filename string, reader io.Reader, size int64) (*data.ImageUploadData, error) {
	if size > maxImageUploadBytes {
		return nil, fmt.Errorf("file too large: max %d bytes", maxImageUploadBytes)
	}
	payload, err := io.ReadAll(io.LimitReader(reader, maxImageUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxImageUploadBytes {
		return nil, fmt.Errorf("file too large: max %d bytes", maxImageUploadBytes)
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	contentType := http.DetectContentType(payload)
	if ct := mime.TypeByExtension(strings.ToLower(path.Ext(filename))); strings.HasPrefix(ct, "image/") {
		contentType = ct
	}
	ext, ok := allowedImageContentTypes[contentType]
	if !ok {
		// DetectContentType may return image/jpeg without matching filename; normalize aliases.
		switch contentType {
		case "image/jpg":
			ext = ".jpg"
			contentType = "image/jpeg"
		default:
			return nil, fmt.Errorf("unsupported image type: %s", contentType)
		}
	}

	sum := md5.Sum(payload)
	hash := hex.EncodeToString(sum[:])
	now := time.Now().UTC()
	prefix := "campaign-center"
	if config.Config != nil && config.Config.OSSConfig != nil && strings.TrimSpace(config.Config.OSSConfig.KeyPrefix) != "" {
		prefix = strings.Trim(strings.TrimSpace(config.Config.OSSConfig.KeyPrefix), "/")
	}
	key := fmt.Sprintf("%s/landing-pages/%04d/%02d/%s%s", prefix, now.Year(), int(now.Month()), hash, ext)

	log.WithContext(ctx).Infow("image_upload", "key", key, "content_type", contentType, "size", len(payload))
	if err := s.storage.PutObject(ctx, key, bytes.NewReader(payload), int64(len(payload)), contentType); err != nil {
		log.WithContext(ctx).Errorw("image_upload_error", "error", err)
		return nil, err
	}
	return &data.ImageUploadData{URL: s.storage.PublicURL(key)}, nil
}
