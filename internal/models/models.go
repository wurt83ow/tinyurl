package models

type Key string

// Request describes the user's request.
type Request struct {
	URL string `json:"url"`
}

// Response describes the server's response.
type Response struct {
	Result string `json:"result"`
}

type RequestURL struct {
	UUID        string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type ResponseURL struct {
	UUID     string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type ResponseUserURL struct {
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
}

type DataURL struct {
	UUID        string `db:"correlation_id" json:"result"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
	UserID      string `db:"user_id" json:"user_id"`
	DeletedFlag bool   `db:"is_deleted" json:"is_deleted"`
}

type DataUser struct {
	UUID  string `db:"id" json:"user_id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Hash  []byte `db:"hash"`
}

type RequestUser struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
type ResponseUser struct {
	Response string `json:"response,omitempty"`
}

type RequestUserReg struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type DeleteURL struct {
	UserID    string   `db:"user_id" json:"user_id"`
	ShortURLs []string `db:"short_url" json:"short_url"`
}
