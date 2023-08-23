package models

// Request описывает запрос пользователя.
type Request struct {
	URL string `json:"url"`
}

// Response описывает ответ сервера.
type Response struct {
	Result string `json:"result"`
}
