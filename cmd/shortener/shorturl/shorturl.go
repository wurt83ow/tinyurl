package shorturl

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// calculate the string hash from hash uint64
func strHash(n uint64) string {
	s := ""
	for n > 0 {
		s = s + alphabet[n%62:(n%62)+1]
		n = n / 62
	}
	return s
}

// calculating big uint64 value of the string
func strToUint64(str string) uint64 {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(str))
	hexstr := hex.EncodeToString(h.Sum(nil))

	bi.SetString(hexstr, 16)
	return bi.Uint64()
}

// calculate the short url from url
func Shorten(url string, shortURLAdress string) (string, string) {
	key := strHash(strToUint64(strings.TrimSpace(url)))
	shortURLAdress = strings.TrimSpace(shortURLAdress)
	if string(shortURLAdress[len(shortURLAdress)-1]) != "/" {
		shortURLAdress += "/"
	}
	fmt.Println(shortURLAdress)
	return key, shortURLAdress + key
}
