package models

// Request описывает запрос пользователя.
type Request struct {
	URL string `json:"url"`
}

// Response описывает ответ сервера.
type Response struct {
	Result string `json:"result"`
}

type RequestRecord struct {
	UUID        string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type ResponseRecord struct {
	UUID     string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type DataURL struct {
	UUID        string `db:"correlation_id" json:"result"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
}
