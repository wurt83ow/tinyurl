package shorturl_test

import (
	"fmt"

	"github.com/wurt83ow/tinyurl/internal/services/shorturl"
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

// // ExamplestrToUint64 demonstrates how to use the strToUint64 function.
// func ExamplestrToUint64() {
// 	str := "example_string"
// 	result := strToUint64(str)
// 	fmt.Printf("String: %s\n", str)
// 	fmt.Printf("Hash as Uint64: %d\n", result)
// }

// // ExamplestrHash demonstrates how to use the strHash function.
// func ExamplestrHash() {
// 	hashValue := uint64(123456789)
// 	result := strHash(hashValue)
// 	fmt.Printf("Hash Value: %d\n", hashValue)
// 	fmt.Printf("Shortened Key: %s\n", result)
// }
