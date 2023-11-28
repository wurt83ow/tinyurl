package shorturl

import (
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BenchmarkLoad(b *testing.B) {

	b.Run("ShortenURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key, shortURLAdress := Shorten(RandStringRunes(8), "http://localhost:8080")

			blackhole := key
			blackhole1 := shortURLAdress
			_, _ = blackhole, blackhole1
		}
	})

}
