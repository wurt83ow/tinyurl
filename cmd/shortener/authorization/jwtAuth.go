package authz

import (
	"crypto/sha256"
	"errors"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
)

var jwtSigningKey []byte
var defaultCookie http.Cookie

var jwtSigningMethod = jwt.SigningMethodHS256

type CustomClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func init() {
	jwtSigningKey = []byte(config.GetAsString("JWT_SIGNING_KEY", "test_key"))
	defaultCookie = http.Cookie{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   config.GetAsString("COOKIE_DOMAIN", "localhost"),
		Secure:   config.GetAsBool("COOKIE_SECURE", true),
	}

}

func CreateJWTTokenForUser(userid string) string {
	claims := CustomClaims{
		userid,
		jwt.StandardClaims{},
	}

	// Encode to token string
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSigningKey)
	if err != nil {
		log.Println("Error occurred generating JWT", err)
		return ""
	}
	return tokenString
}

func DecodeJWTToUser(token string) (string, error) {
	// Decode
	decodeToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if !(jwtSigningMethod == token.Method) {
			// Check our method hasn't changed since issuance
			return nil, errors.New("signing method mismatch")
		}
		return jwtSigningKey, nil
	})

	// There's two parts. We might decode it successfully but it might
	// be the case we aren't Valid so you must check both
	if decClaims, ok := decodeToken.Claims.(*CustomClaims); ok && decodeToken.Valid {
		return decClaims.Email, nil
	}
	return "", err
}

func GetHash(email string, password string) []byte {
	src := []byte(email + password)

	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	// передаём байты для хеширования
	h.Write(src)
	// вычисляем хеш
	return h.Sum(nil)

}
func AuthCookie(token string) *http.Cookie {
	d := defaultCookie
	d.Name = "jwt-token"
	d.Value = token
	d.Path = "/"
	return &d
}
