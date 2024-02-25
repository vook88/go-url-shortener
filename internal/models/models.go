package models

type RequestShortURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	ShortURL string `json:"result"`
}

type RequestBatchLongURLs []BatchLongURL

type BatchLongURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatchShortURLs []BatchShortURL

type BatchShortURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchUserURLs []UserURL
