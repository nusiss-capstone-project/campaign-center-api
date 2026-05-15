package service

import (
	"context"
)

func (s *landingPageTranslationService) ListTranslatedLangs(
	ctx context.Context, landingPageID int64,
) ([]string, error) {
	_ = ctx
	if _, err := s.pages.GetByID(landingPageID); err != nil {
		return nil, err
	}
	return s.translations.ListLangsByLandingPageID(landingPageID)
}
