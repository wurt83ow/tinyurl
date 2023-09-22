package models

type Key string

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

type ResponseUserURLs struct {
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
}

type DataURL struct {
	UUID        string `db:"correlation_id" json:"result"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
	UserID      string `db:"user_id" json:"user_id"`
}

type DataUser struct {
	UUID  string `db:"id" json:"user_id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Hash  []byte `db:"hash"`
}

type RequestBody struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
type ResponseBody struct {
	Response string `json:"response,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
