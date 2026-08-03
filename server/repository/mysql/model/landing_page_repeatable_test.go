package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLandingPageRepeatableItems_ValueAndScan(t *testing.T) {
	items := LandingPageRepeatableItems{{Title: "a", Description: "b"}}
	v, err := items.Value()
	require.NoError(t, err)
	require.Equal(t, `[{"title":"a","description":"b"}]`, v)

	var scanned LandingPageRepeatableItems
	require.NoError(t, scanned.Scan(v))
	require.Equal(t, items, scanned)

	require.NoError(t, scanned.Scan(nil))
	require.Empty(t, scanned)

	require.NoError(t, scanned.Scan([]byte(`[{"title":"x","description":"y"}]`)))
	require.Equal(t, "x", scanned[0].Title)

	require.Error(t, scanned.Scan(123))
	require.Error(t, (*LandingPageRepeatableItems)(nil).Scan([]byte("[]")))
}

func TestLandingPageRepeatableItems_ValueNilSlice(t *testing.T) {
	var items LandingPageRepeatableItems
	v, err := items.Value()
	require.NoError(t, err)
	require.Equal(t, "[]", v)
}
