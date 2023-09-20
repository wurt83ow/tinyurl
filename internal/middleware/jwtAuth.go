package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
)

var jwtSigningKey []byte
var defaultCookie http.Cookie
var jwtSessionLength time.Duration
var jwtSigningMethod = jwt.SigningMethodHS256

func init() {
	jwtSigningKey = []byte(config.GetAsString("JWT_SIGNING_KEY", "test_key"))
	defaultCookie = http.Cookie{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   config.GetAsString("COOKIE_DOMAIN", "localhost"),
		Secure:   config.GetAsBool("COOKIE_SECURE", true),
	}
	jwtSessionLength = time.Duration(config.GetAsInt("JWT_SESSION_LENGTH", 50))
}

// JWTProtectedMiddleware verifies a valid JWT exists in our
// cookie and if not, encourages the consumer to login again.
func JWTProtectedMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Grab jwt-token cookie
		jwtCookie, err := r.Cookie("jwt-token")

		if err != nil {
			log.Println("Error occurred reading cookie", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message,omitempty"`
			}{
				Message: "Your session is not valid - please login",
			})

			return
		}

		log.Println("Got cookie value", jwtCookie.Value)

		// Decode and validate JWT if there is one
		email, err := DecodeJWTToUser(jwtCookie.Value)

		if email == "" || err != nil {
			log.Println("Error decoding token", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message,omitempty"`
			}{
				Message: "Your session is not valid - please login",
			})
			return
		}

		// If it's good, update the expiry time
		freshToken := CreateJWTTokenForUser(email)

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", email)

		// Set the new cookie and continue into the handler
		w.Header().Add("Content-Type", "application/json")
		http.SetCookie(w, AuthCookie(freshToken))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type CustomClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func CreateJWTTokenForUser(userid string) string {
	claims := CustomClaims{
		userid,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * jwtSessionLength).Unix(),
		},
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

func ExpiredAuthCookie() *http.Cookie {
	d := defaultCookie
	d.Name = "jwt-token"
	d.Value = ""
	d.Path = "/"
	d.MaxAge = -1
	d.Expires = time.Date(1983, 7, 26, 20, 34, 58, 651387237, time.UTC)
	return &d
}
