package mysql

import "github.com/lianjin/campaign-center-api/server/repository/mysql/model"

type noopLandingPageTranslationRepository struct{}

func (noopLandingPageTranslationRepository) GetByLandingPageAndLang(
	int64, string,
) (*model.CampaignLandingPageTranslation, error) {
	return nil, nil
}

func (noopLandingPageTranslationRepository) ListLangsByLandingPageID(int64) ([]string, error) {
	return []string{}, nil
}

func (noopLandingPageTranslationRepository) Upsert(*model.CampaignLandingPageTranslation) error {
	return nil
}

// NewNoopLandingPageTranslationRepository returns an in-memory no-op repo for tests.
func NewNoopLandingPageTranslationRepository() LandingPageTranslationRepository {
	return noopLandingPageTranslationRepository{}
}
