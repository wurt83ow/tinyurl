// Package authz provides authentication and authorization functionality, including JWT token handling.
package authz

import (
	"context"
	"crypto/sha256"
	"errors"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Storage is an interface representing methods for inserting user data.
type Storage interface {
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
}

// CustomClaims represents custom claims for JWT token.
type CustomClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Log is an interface representing a logger with Info method.
type Log interface {
	Info(string, ...zapcore.Field)
}

// JWTAuthz provides JWT token creation, decoding, and middleware functionality for authentication and authorization.
type JWTAuthz struct {
	jwtSigningKey    []byte
	log              Log
	jwtSigningMethod *jwt.SigningMethodHMAC
	defaultCookie    http.Cookie
}

// NewJWTAuthz creates a new JWTAuthz instance with the provided signing key and logger.
func NewJWTAuthz(signingKey string, log Log) *JWTAuthz {
	return &JWTAuthz{
		jwtSigningKey:    []byte(config.GetAsString("JWT_SIGNING_KEY", signingKey)),
		log:              log,
		jwtSigningMethod: jwt.SigningMethodHS256,

		defaultCookie: http.Cookie{
			HttpOnly: true,
		},
	}
}

// JWTAuthzMiddleware returns a middleware function that verifies the presence of a valid JWT in the cookie.
// If not present, it encourages the user to log in again.
func (j *JWTAuthz) JWTAuthzMiddleware(storage Storage, log Log) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Grab jwt-token cookie
			jwtCookie, err := r.Cookie("jwt-token")

			var userID string
			if err == nil {
				userID, err = j.DecodeJWTToUser(jwtCookie.Value)
				if err != nil {
					userID = ""
					log.Info("Error occurred creating a cookie", zap.Error(err))
				}
			} else {
				log.Info("Error occurred reading cookie", zap.Error(err))
			}

			if userID == "" {
				jwtCookie := r.Header.Get("Authorization")

				if jwtCookie != "" {
					userID, err = j.DecodeJWTToUser(jwtCookie)

					if err != nil {
						userID = ""
						log.Info("Error occurred creating a cookie", zap.Error(err))
					}
				}
			}

			if userID == "" {
				userID = uuid.New().String()

				go func() {
					email := uuid.New().String()
					dataUser := models.DataUser{UUID: userID, Email: email, Name: "default"}
					_, err = storage.InsertUser(email, dataUser)
					if err != nil {
						log.Info("Error occurred user create", zap.Error(err))
					}
				}()

				freshToken := j.CreateJWTTokenForUser(userID)
				http.SetCookie(w, j.AuthCookie("jwt-token", freshToken))
				w.Header().Set("Authorization", freshToken)
			}

			var keyUserID models.Key = "userID"
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

// CreateJWTTokenForUser creates a JWT token for the specified user ID.
func (j *JWTAuthz) CreateJWTTokenForUser(userid string) string {
	claims := CustomClaims{
		userid,
		jwt.StandardClaims{},
	}

	// Encode to token string
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.jwtSigningKey)
	if err != nil {
		log.Println("Error occurred generating JWT", err)
		return ""
	}

	return tokenString
}

// DecodeJWTToUser decodes a JWT token to retrieve the user ID.
func (j *JWTAuthz) DecodeJWTToUser(token string) (string, error) {
	// Decode
	decodeToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if !(j.jwtSigningMethod == token.Method) {
			// Check our method hasn't changed since issuance
			return nil, errors.New("signing method mismatch")
		}
		return j.jwtSigningKey, nil
	})

	// There's two parts. We might decode it successfully but it might
	// be the case we aren't Valid so you must check both
	if decClaims, ok := decodeToken.Claims.(*CustomClaims); ok && decodeToken.Valid {
		return decClaims.Email, nil
	}

	return "", err
}

// GetHash computes the SHA-256 hash of the concatenation of email and password.
func (j *JWTAuthz) GetHash(email string, password string) []byte {
	src := []byte(email + password)

	// create a new hash.Hash that calculates the SHA-256 checksum
	h := sha256.New()
	// transfer bytes for hashing
	h.Write(src)
	// calculate the hash

	return h.Sum(nil)
}

// AuthCookie creates an http.Cookie with the specified name and token value.
func (j *JWTAuthz) AuthCookie(name string, token string) *http.Cookie {
	d := j.defaultCookie
	d.Name = name
	d.Value = token
	d.Path = "/"

	return &d
}
