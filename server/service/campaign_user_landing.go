package service

import (
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

func (s *userCampaignService) resolveLandingPageTexts(
	lp *model.CampaignLandingPage, lang string,
) (title, desc, terms, resolvedLang string, err error) {
	if lang == "" || lang == lp.DefaultLang {
		return lp.Title, lp.Description, lp.Terms, lp.DefaultLang, nil
	}
	tr, err := s.translations.GetByLandingPageAndLang(lp.ID, lang)
	if err != nil {
		return "", "", "", "", err
	}
	if tr != nil {
		return tr.Title, tr.Description, tr.Terms, lang, nil
	}
	return lp.Title, lp.Description, lp.Terms, lp.DefaultLang, nil
}
