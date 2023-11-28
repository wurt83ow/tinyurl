// Package models provides data structures used in the application.
package models

// Key is an alias for string and represents a key used in various contexts.
type Key string

// Request describes the user's request.
type Request struct {
	URL string `json:"url"`
}

// Response describes the server's response.
type Response struct {
	Result string `json:"result"`
}

// DataURLite represents a simplified version of data related to a URL.
type DataURLite struct {
	UUID        string `db:"correlation_id" json:"correlation_id"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
}

// DataURL represents data related to a URL.
type DataURL struct {
	UUID        string `db:"correlation_id" json:"result"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
	UserID      string `db:"user_id" json:"user_id"`
	DeletedFlag bool   `db:"is_deleted" json:"is_deleted"`
}

// DataUser represents data related to a user.
type DataUser struct {
	UUID  string `db:"id" json:"user_id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Hash  []byte `db:"hash"`
}

// RequestUser describes the user's request for user-related operations.
type RequestUser struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// ResponseUser describes the server's response for user-related operations.
type ResponseUser struct {
	Response string `json:"response,omitempty"`
}

// DeleteURL represents a request to delete one or more URLs associated with a user.
type DeleteURL struct {
	UserID    string   `db:"user_id" json:"user_id"`
	ShortURLs []string `db:"short_url" json:"short_url"`
}
