// Package shorturl provides functionality for shortening URLs.
package shorturl

import (
	"crypto/md5"
	"encoding/hex"
	"math/big"
	"strings"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// strHash calculates the string hash from a uint64 value.
func strHash(n uint64) string {
	var s string
	for n > 0 {
		s = s + alphabet[n%62:(n%62)+1]
		n = n / 62
	}

	return s
}

// strToUint64 calculates the big uint64 value of the string.
func strToUint64(str string) uint64 {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(str))
	hexstr := hex.EncodeToString(h.Sum(nil))

	bi.SetString(hexstr, 16)

	return bi.Uint64()
}

// Shorten generates a short URL from the given URL and base short URL address.
func Shorten(url string, shortURLAdress string) (string, string) {
	key := strHash(strToUint64(strings.TrimSpace(url)))
	shortURLAdress = strings.TrimSpace(shortURLAdress)
	if string(shortURLAdress[len(shortURLAdress)-1]) != "/" {
		shortURLAdress += "/"
	}

	return key, shortURLAdress + key

}
