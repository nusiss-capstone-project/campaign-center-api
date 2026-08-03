package service

import (
	"strings"
	"testing"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/stretchr/testify/require"
)

func TestValidateLandingPageContent_ok(t *testing.T) {
	err := validateLandingPageContent(
		[]data.LandingPageRepeatableItemVO{{Title: "s", Description: "d"}},
		[]data.LandingPageRepeatableItemVO{},
	)
	require.NoError(t, err)
}

func TestValidateLandingPageContent_tooManySteps(t *testing.T) {
	items := make([]data.LandingPageRepeatableItemVO, maxLandingPageRepeatableItems+1)
	for i := range items {
		items[i] = data.LandingPageRepeatableItemVO{Title: "t", Description: "d"}
	}
	err := validateLandingPageContent(items, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at most")
}

func TestValidateLandingPageContent_requiresTitleAndDescription(t *testing.T) {
	err := validateLandingPageContent([]data.LandingPageRepeatableItemVO{{Title: " ", Description: "d"}}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "title")

	err = validateLandingPageContent([]data.LandingPageRepeatableItemVO{{Title: "t", Description: ""}}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "description")
}

func TestRepeatableConverters(t *testing.T) {
	require.Empty(t, toModelRepeatableItems(nil))
	require.Empty(t, toDataRepeatableItems(nil))
	require.Empty(t, normalizeRepeatableVO(nil))

	in := []data.LandingPageRepeatableItemVO{{Title: "a", Description: "b"}}
	m := toModelRepeatableItems(in)
	require.Equal(t, model.LandingPageRepeatableItems{{Title: "a", Description: "b"}}, m)
	require.Equal(t, in, toDataRepeatableItems(m))
	require.Equal(t, in, normalizeRepeatableVO(in))
}

func TestLandingPageCreateUpdateResp(t *testing.T) {
	body := data.LandingPageBody{
		DefaultLang: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
		Steps: []data.LandingPageRepeatableItemVO{{Title: "s", Description: "sd"}},
	}
	created := landingPageCreateResp(9, 1, body)
	require.Equal(t, int64(9), created.LandingPageID)
	require.Equal(t, int16(1), created.Status)
	require.Equal(t, "t", created.Title)
	require.Len(t, created.Steps, 1)

	updated := landingPageUpdateResp(9, body)
	require.Equal(t, int64(9), updated.LandingPageID)
	require.True(t, strings.HasPrefix(updated.Title, "t"))
}
