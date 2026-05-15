package service

import (
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// LandingPageDetailView is resolved landing page text for admin or user display.
type LandingPageDetailView struct {
	ID             int64
	Lang           string
	DefaultLang    string
	BannerImageURL string
	Title          string
	Description    string
	Terms          string
	Status         int16
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func landingPageViewFromRow(p *model.CampaignLandingPage) *LandingPageDetailView {
	return &LandingPageDetailView{
		ID:             p.ID,
		DefaultLang:    p.DefaultLang,
		BannerImageURL: p.BannerImageURL,
		Title:          p.Title,
		Description:    p.Description,
		Terms:          p.Terms,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func applyTranslationToView(v *LandingPageDetailView, lang, title, description, terms string) {
	v.Lang = lang
	v.Title = title
	v.Description = description
	v.Terms = terms
}
