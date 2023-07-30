package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/wurt83ow/tinyurl/cmd/shortener/shortener"
)

const (
	CONN_HOST = ""
	CONN_PORT = "8080"
)

// HANDLERS

// POST
func shortenURL(w http.ResponseWriter, r *http.Request) {
	// установим правильный заголовок для типа данных
	body, _ := io.ReadAll(r.Body)
	w.Header().Set("Content-Type", "text/plain")
	// пока установим ответ-заглушку, без проверки ошибок
	_, _ = w.Write(body)
	// Получить сокращенную ссылку или вывести в лог ошибку и произвести какие-то (?) действия
	// Сохранить сокращенную ссылку и полную ссылку в хранилище или вывести в лог ошибку и произвести какие-то (?) действия
	// Дать ответ клиенту (сокр.url text/plain) и код ответа (201) (может быть ошибка здесь?)или уведомить об ошибке (ответ с кодом 400.).

}

// GET
func getFullUrl(w http.ResponseWriter, r *http.Request) {
	// log.Printf("Metod GET ")
	w.Write([]byte(r.URL.Path))
	//Получить сокр.url в виде (/EwHXdJfB). // Нужны доп.действия?
	// Получить полную ссылку из хранилища или вывести в лог ошибку и произвести какие-то (?) действия
	// Дать ответ клиенту: код ответа (307) и оригинальный URL в заголовке Location ИЛИ уведомить об ошибке (ответ с кодом 400.)
	//
	// jsonBody, err := json.Marshal(results)
	// if err != nil {
	// 	http.Error(w, "Error converting results to json",
	// 		http.StatusInternalServerError)
	// }
	// w.Write(jsonBody)
}

// HTTP-SERVER

func webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		shortenURL(w, r)
	} else if r.Method == http.MethodGet {
		getFullUrl(w, r)
	} else {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func run() error {
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(webhook))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}

	fmt.Println(shortener.ShortUrl(shortener.StrToUint64("https://practicum.yandex.ru/")))
}
