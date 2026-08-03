package proxy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotConfiguredObjectStorage(t *testing.T) {
	s := notConfiguredObjectStorage{}
	require.ErrorIs(t, s.PutObject(context.Background(), "k", nil, 0, ""), ErrOSSNotConfigured)
	require.Empty(t, s.PublicURL("k"))
}

func TestAliyunOSSStorage_PublicURL(t *testing.T) {
	withCDN := &aliyunOSSStorage{
		publicBaseURL: "https://cdn.example",
		endpoint:      "oss-ap-southeast-1.aliyuncs.com",
		bucketName:    "bucket",
	}
	require.Equal(t, "https://cdn.example/a/b.png", withCDN.PublicURL("/a/b.png"))

	fallback := &aliyunOSSStorage{
		endpoint:   "oss-ap-southeast-1.aliyuncs.com",
		bucketName: "bucket",
	}
	require.Equal(t,
		"https://bucket.oss-ap-southeast-1.aliyuncs.com/a/b.png",
		fallback.PublicURL("a/b.png"),
	)
}
