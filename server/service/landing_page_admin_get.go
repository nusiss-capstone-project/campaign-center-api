package service

func (s *landingPageAdminService) GetLandingPage(id int64, lang string) (*LandingPageDetailView, error) {
	page, err := s.pages.GetByID(id)
	if err != nil {
		return nil, err
	}
	view := landingPageViewFromRow(page)
	if lang == "" || lang == page.DefaultLang {
		view.Lang = page.DefaultLang
		return view, nil
	}
	tr, err := s.translations.GetByLandingPageAndLang(id, lang)
	if err != nil {
		return nil, err
	}
	if tr != nil {
		applyTranslationToView(view, lang, tr.Title, tr.Description, tr.Terms)
		return view, nil
	}
	view.Lang = page.DefaultLang
	return view, nil
}
