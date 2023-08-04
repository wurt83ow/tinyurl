package shorturl

import (
	"testing"
)

func TestShorten(t *testing.T) {
	tests := []struct { // добавляем слайс тестов
		name  string
		url   string
		proto string
		host  string

		key   string
		shurl string
	}{
		{
			name:  "simple test #1",
			url:   "https://practicum.yandex.ru/",
			proto: "http",
			host:  "localhost:8080",
			key:   "nOykhckC3Od",
			shurl: "http://localhost:8080/nOykhckC3Od",
		},
	}
	for _, test := range tests { // цикл по всем тестам
		t.Run(test.name, func(t *testing.T) {
			if key, shurl := Shorten(test.url, test.proto, test.host); key != test.key || shurl != test.shurl {
				t.Errorf(shurl)
			}
		})
	}
}
