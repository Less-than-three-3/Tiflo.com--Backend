package model

type TranslationResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	QuotaFinished   bool   `json:"quotaFinished"`
	ResponseDetails string `json:"responseDetails"`
}
