package data

// LandingPageRepeatableItemVO is one steps/faq entry.
type LandingPageRepeatableItemVO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// LandingPageBody is the create/update landing page request body.
type LandingPageBody struct {
	DefaultLang    string                       `json:"defaultLang" binding:"required"`
	BannerImageURL string                       `json:"bannerImageUrl" binding:"required"`
	Title          string                       `json:"title" binding:"required"`
	Description    string                       `json:"description" binding:"required"`
	Terms          string                       `json:"terms" binding:"required"`
	Steps          []LandingPageRepeatableItemVO `json:"steps"`
	Faq            []LandingPageRepeatableItemVO `json:"faq"`
}

// LandingPageCreateResp is returned after create (submitted fields + id/status).
type LandingPageCreateResp struct {
	LandingPageID  int64                         `json:"landingPageId"`
	Status         int16                         `json:"status"`
	DefaultLang    string                        `json:"defaultLang"`
	BannerImageURL string                        `json:"bannerImageUrl"`
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	Terms          string                        `json:"terms"`
	Steps          []LandingPageRepeatableItemVO `json:"steps"`
	Faq            []LandingPageRepeatableItemVO `json:"faq"`
}

// LandingPageUpdateResp is returned after update (submitted fields + id).
type LandingPageUpdateResp struct {
	LandingPageID  int64                         `json:"landingPageId"`
	DefaultLang    string                        `json:"defaultLang"`
	BannerImageURL string                        `json:"bannerImageUrl"`
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	Terms          string                        `json:"terms"`
	Steps          []LandingPageRepeatableItemVO `json:"steps"`
	Faq            []LandingPageRepeatableItemVO `json:"faq"`
}

// LandingPageListItemVO is a compact list row.
type LandingPageListItemVO struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status int16  `json:"status"`
}

// LandingPageListData is the list response data envelope.
type LandingPageListData struct {
	Total int64                  `json:"total"`
	Items []LandingPageListItemVO `json:"items"`
}

// LandingPageDetailVO is Get/locale-detail response data.
type LandingPageDetailVO struct {
	ID             int64                         `json:"id"`
	Lang           string                        `json:"lang"`
	DefaultLang    string                        `json:"defaultLang"`
	BannerImageURL string                        `json:"bannerImageUrl"`
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	Terms          string                        `json:"terms"`
	Steps          []LandingPageRepeatableItemVO `json:"steps"`
	Faq            []LandingPageRepeatableItemVO `json:"faq"`
	Status         int16                         `json:"status"`
	CreatedAt      string                        `json:"createdAt"`
	UpdatedAt      string                        `json:"updatedAt"`
}

// LandingPagePublishResp is returned after publish.
type LandingPagePublishResp struct {
	LandingPageID int64 `json:"landingPageId"`
	Status        int16 `json:"status"`
}

// LandingPageTranslationVO is a persisted translation row shape.
type LandingPageTranslationVO struct {
	ID            int64                         `json:"id"`
	LandingPageID int64                         `json:"landingPageId"`
	Lang          string                        `json:"lang"`
	Title         string                        `json:"title"`
	Description   string                        `json:"description"`
	Terms         string                        `json:"terms"`
	Steps         []LandingPageRepeatableItemVO `json:"steps"`
	Faq           []LandingPageRepeatableItemVO `json:"faq"`
	UpdatedAt     int64                         `json:"updatedAt"`
	CreatedAt     int64                         `json:"createdAt"`
	CreatedBy     string                        `json:"createdBy"`
	UpdatedBy     string                        `json:"updatedBy"`
}

// GenerateLandingTranslationReq is the JSON body for machine translation preview.
type GenerateLandingTranslationReq struct {
	SourceLang  string                        `json:"sourceLang" binding:"required" example:"en"`
	TargetLang  string                        `json:"targetLang" binding:"required" example:"ja"`
	Title       string                        `json:"title"`
	Description string                        `json:"description"`
	Terms       string                        `json:"terms"`
	Steps       []LandingPageRepeatableItemVO `json:"steps"`
	Faq         []LandingPageRepeatableItemVO `json:"faq"`
}

// GenerateLandingTranslationData is StandardResponse.data after generate.
type GenerateLandingTranslationData struct {
	Lang        string                        `json:"lang" example:"ja"`
	Title       string                        `json:"title"`
	Description string                        `json:"description"`
	Terms       string                        `json:"terms"`
	Steps       []LandingPageRepeatableItemVO `json:"steps"`
	Faq         []LandingPageRepeatableItemVO `json:"faq"`
}

// PutLandingTranslationReq is the JSON body for upserting one locale.
type PutLandingTranslationReq struct {
	Title       string                        `json:"title" binding:"required"`
	Description string                        `json:"description" binding:"required"`
	Terms       string                        `json:"terms" binding:"required"`
	Steps       []LandingPageRepeatableItemVO `json:"steps"`
	Faq         []LandingPageRepeatableItemVO `json:"faq"`
	Operator    string                        `json:"operator" example:"admin"`
}

// PutLandingTranslationData is StandardResponse.data after save.
type PutLandingTranslationData struct {
	LandingPageID int64  `json:"landingPageId"`
	Lang          string `json:"lang"`
}

// LandingPageTranslatedLangsData is StandardResponse.data for persisted translation locales.
type LandingPageTranslatedLangsData struct {
	Langs []string `json:"langs"`
}

// ImageUploadData is StandardResponse.data after image upload.
type ImageUploadData struct {
	URL string `json:"url"`
}
