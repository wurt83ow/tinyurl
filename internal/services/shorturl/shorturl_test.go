package shorturl

import (
	"testing"
)

// shortURLAdress := configs.ShortURLAdress()
func TestShorten(t *testing.T) {
	tests := []struct { // adding slice tests
		name           string
		url            string
		shortURLAdress string
		key            string
		shurl          string
	}{
		{
			name:           "simple test #1",
			url:            "https://practicum.yandex.ru/",
			shortURLAdress: "http://localhost:8080/",
			key:            "nOykhckC3Od",
			shurl:          "http://localhost:8080/nOykhckC3Od",
		},
	}
	for _, test := range tests { // cycle through all tests
		t.Run(test.name, func(t *testing.T) {
			if key, shurl := Shorten(test.url, test.shortURLAdress); key != test.key || shurl != test.shurl {
				t.Errorf(shurl)
			}
		})
	}
}
