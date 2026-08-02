package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

const maxLandingPageRepeatableItems = 10

func validateLandingPageRepeatableItems(field string, items []data.LandingPageRepeatableItemVO) error {
	if len(items) > maxLandingPageRepeatableItems {
		return fmt.Errorf("%s: at most %d items allowed", field, maxLandingPageRepeatableItems)
	}
	for i, item := range items {
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("%s[%d].title is required", field, i)
		}
		if strings.TrimSpace(item.Description) == "" {
			return fmt.Errorf("%s[%d].description is required", field, i)
		}
	}
	return nil
}

func validateLandingPageContent(steps, faq []data.LandingPageRepeatableItemVO) error {
	if err := validateLandingPageRepeatableItems("steps", steps); err != nil {
		return err
	}
	return validateLandingPageRepeatableItems("faq", faq)
}

func toModelRepeatableItems(items []data.LandingPageRepeatableItemVO) model.LandingPageRepeatableItems {
	if items == nil {
		return model.LandingPageRepeatableItems{}
	}
	out := make(model.LandingPageRepeatableItems, 0, len(items))
	for _, item := range items {
		out = append(out, model.LandingPageRepeatableItem{
			Title:       item.Title,
			Description: item.Description,
		})
	}
	return out
}

func toDataRepeatableItems(items model.LandingPageRepeatableItems) []data.LandingPageRepeatableItemVO {
	if items == nil {
		return []data.LandingPageRepeatableItemVO{}
	}
	out := make([]data.LandingPageRepeatableItemVO, 0, len(items))
	for _, item := range items {
		out = append(out, data.LandingPageRepeatableItemVO{
			Title:       item.Title,
			Description: item.Description,
		})
	}
	return out
}

func normalizeRepeatableVO(items []data.LandingPageRepeatableItemVO) []data.LandingPageRepeatableItemVO {
	if items == nil {
		return []data.LandingPageRepeatableItemVO{}
	}
	return items
}

func landingPageCreateResp(id int64, status int16, body data.LandingPageBody) *data.LandingPageCreateResp {
	return &data.LandingPageCreateResp{
		LandingPageID:  id,
		Status:         status,
		DefaultLang:    body.DefaultLang,
		BannerImageURL: body.BannerImageURL,
		Title:          body.Title,
		Description:    body.Description,
		Terms:          body.Terms,
		Steps:          normalizeRepeatableVO(body.Steps),
		Faq:            normalizeRepeatableVO(body.Faq),
	}
}

func landingPageUpdateResp(id int64, body data.LandingPageBody) *data.LandingPageUpdateResp {
	return &data.LandingPageUpdateResp{
		LandingPageID:  id,
		DefaultLang:    body.DefaultLang,
		BannerImageURL: body.BannerImageURL,
		Title:          body.Title,
		Description:    body.Description,
		Terms:          body.Terms,
		Steps:          normalizeRepeatableVO(body.Steps),
		Faq:            normalizeRepeatableVO(body.Faq),
	}
}

func landingPageDetailVO(v *LandingPageDetailView) *data.LandingPageDetailVO {
	return &data.LandingPageDetailVO{
		ID:             v.ID,
		Lang:           v.Lang,
		DefaultLang:    v.DefaultLang,
		BannerImageURL: v.BannerImageURL,
		Title:          v.Title,
		Description:    v.Description,
		Terms:          v.Terms,
		Steps:          normalizeRepeatableVO(v.Steps),
		Faq:            normalizeRepeatableVO(v.Faq),
		Status:         v.Status,
		CreatedAt:      v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      v.UpdatedAt.Format(time.RFC3339),
	}
}
