package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	authz "github.com/wurt83ow/tinyurl/cmd/shortener/authorization"
	"github.com/wurt83ow/tinyurl/internal/models"
)

type Storage interface {
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
}

// JWTProtectedMiddleware verifies a valid JWT exists in our
// cookie and if not, encourages the consumer to login again.
func JWTProtectedMiddleware(storage Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Grab jwt-token cookie
			jwtCookie, err := r.Cookie("jwt-token")

			userID := ""
			if err == nil {
				userID, err = authz.DecodeJWTToUser(jwtCookie.Value)
				if err != nil {
					userID = ""
					log.Println("Error occurred creating a cookie", err)
				}

			} else {
				log.Println("Error occurred reading cookie", err)
			}

			if userID == "" {
				jwtCookie := r.Header.Get("Authorization")

				if jwtCookie != "" {
					userID, err = authz.DecodeJWTToUser(jwtCookie)
					if err != nil {
						userID = ""
						log.Println("Error occurred creating a cookie", err)
					}
				}

			}

			if userID == "" {

				// save full url to storage with the key received earlier
				email := uuid.New().String()
				userID = uuid.New().String()
				dataUser := models.DataUser{UUID: userID, Email: email, Name: "default"}
				_, err = storage.InsertUser(email, dataUser)

				freshToken := authz.CreateJWTTokenForUser(userID)
				http.SetCookie(w, authz.AuthCookie(freshToken))
				w.Header().Set("Authorization", freshToken)
				if err != nil {
					log.Println("Error occurred user create", err)
				}
			}

			var keyUserID models.Key = "userID"
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyUserID, userID)

			next.ServeHTTP(w, r.WithContext(ctx))

		}

		return http.HandlerFunc(fn)
	}
}
