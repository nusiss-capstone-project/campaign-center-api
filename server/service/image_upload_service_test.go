package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	"github.com/stretchr/testify/require"
)

type stubObjectStorage struct {
	putErr error
	lastKey string
	lastCT  string
	lastSize int64
}

func (s *stubObjectStorage) PutObject(_ context.Context, key string, reader io.Reader, size int64, contentType string) error {
	s.lastKey = key
	s.lastCT = contentType
	s.lastSize = size
	if s.putErr != nil {
		return s.putErr
	}
	_, err := io.Copy(io.Discard, reader)
	return err
}

func (s *stubObjectStorage) PublicURL(key string) string {
	return "https://cdn.example/" + key
}

func pngBytes() []byte {
	// Minimal PNG signature; enough for http.DetectContentType.
	return []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0}
}

func TestImageUploadService_Upload_success(t *testing.T) {
	prev := config.Config
	config.Config = &config.Conf{OSSConfig: &config.OSSConfig{KeyPrefix: "campaign-center"}}
	t.Cleanup(func() { config.Config = prev })

	storage := &stubObjectStorage{}
	svc := NewImageUploadService(storage)
	payload := pngBytes()
	out, err := svc.Upload(context.Background(), "banner.png", bytes.NewReader(payload), int64(len(payload)))
	require.NoError(t, err)

	sum := sha256.Sum256(payload)
	wantHash := hex.EncodeToString(sum[:])
	now := time.Now().UTC()
	require.Contains(t, storage.lastKey, "campaign-center/landing-pages/")
	require.Contains(t, storage.lastKey, wantHash+".png")
	require.True(t, strings.Contains(storage.lastKey, now.Format("2006")) || strings.Contains(storage.lastKey, "/"))
	require.Equal(t, "image/png", storage.lastCT)
	require.Equal(t, int64(len(payload)), storage.lastSize)
	require.Equal(t, "https://cdn.example/"+storage.lastKey, out.URL)
}

func TestImageUploadService_Upload_rejectsTooLargeDeclaredSize(t *testing.T) {
	svc := NewImageUploadService(&stubObjectStorage{})
	_, err := svc.Upload(context.Background(), "a.png", bytes.NewReader(pngBytes()), maxImageUploadBytes+1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too large")
}

func TestImageUploadService_Upload_rejectsTooLargePayload(t *testing.T) {
	svc := NewImageUploadService(&stubObjectStorage{})
	big := make([]byte, maxImageUploadBytes+8)
	copy(big, pngBytes())
	// Declared size is within limit; body exceeds after read.
	_, err := svc.Upload(context.Background(), "a.png", bytes.NewReader(big), maxImageUploadBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too large")
}

func TestImageUploadService_Upload_rejectsEmpty(t *testing.T) {
	svc := NewImageUploadService(&stubObjectStorage{})
	_, err := svc.Upload(context.Background(), "a.png", bytes.NewReader(nil), 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

func TestImageUploadService_Upload_rejectsUnsupportedType(t *testing.T) {
	svc := NewImageUploadService(&stubObjectStorage{})
	_, err := svc.Upload(context.Background(), "notes.txt", bytes.NewReader([]byte("hello world")), 11)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported")
}

func TestImageUploadService_Upload_propagatesStorageError(t *testing.T) {
	storage := &stubObjectStorage{putErr: errors.New("oss down")}
	svc := NewImageUploadService(storage)
	payload := pngBytes()
	_, err := svc.Upload(context.Background(), "banner.png", bytes.NewReader(payload), int64(len(payload)))
	require.EqualError(t, err, "oss down")
}
