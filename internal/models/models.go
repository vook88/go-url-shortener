package models

type RequestShortURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	ShortURL string `json:"result"`
}
