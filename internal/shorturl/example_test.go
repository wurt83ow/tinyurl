package shorturl_test

import (
	"fmt"

	"github.com/wurt83ow/tinyurl/internal/shorturl"
)

// ExampleShorten demonstrates how to use the Shorten function to generate a short URL.
func ExampleShorten() {
	url := "https://practicum.yandex.ru/"
	baseShortURL := "http://localhost:8080/"
	key, shortURL := shorturl.Shorten(url, baseShortURL)

	fmt.Printf("Shortened Key: %s\n", key)
	fmt.Printf("Shortened URL: %s\n", shortURL)

	// Output:
	// Shortened Key: nOykhckC3Od
	// Shortened URL: http://localhost:8080/nOykhckC3Od
}
