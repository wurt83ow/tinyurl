package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	authz "github.com/wurt83ow/tinyurl/cmd/shortener/authorization"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
)

type Storage interface {
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
}

// JWTAuthzMiddleware verifies a valid JWT exists in our
// cookie and if not, encourages the consumer to login again.
func JWTAuthzMiddleware(storage Storage, log Log) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Grab jwt-token cookie

			jwtCookie, err := r.Cookie("jwt-token")
			// jwtCookie, err := r.Cookie("Authorization")
			fmt.Println("1111111111111111111111111111", r.Cookies())
			// Loop over header names
			for name, values := range r.Header {
				// Loop over all values for the name.
				for _, value := range values {
					fmt.Println("222222222222222222222222222", name, value)
				}
			}
			userID := ""
			if err == nil {
				userID, err = authz.DecodeJWTToUser(jwtCookie.Value)
				if err != nil {
					userID = ""
					log.Info("Error occurred creating a cookie", zap.Error(err))
				}

			} else {
				log.Info("Error occurred reading cookie", zap.Error(err))
			}

			if userID == "" {
				jwtCookie := r.Header.Get("Authorization")
				fmt.Println("JWTAuthzMiddleware header.Authorization", jwtCookie)
				if jwtCookie != "" {
					userID, err = authz.DecodeJWTToUser(jwtCookie)
					fmt.Println("JWTAuthzMiddleware header.Authorization.userID", userID, err)
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

				freshToken := authz.CreateJWTTokenForUser(userID)
				http.SetCookie(w, authz.AuthCookie("jwt-token", freshToken))
				http.SetCookie(w, authz.AuthCookie("Authorization", freshToken))
				w.Header().Set("Authorization", freshToken)
				fmt.Println("func JWTAuthzMiddleware set cookie", freshToken)
			}

			var keyUserID models.Key = "userID"
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyUserID, userID)
			fmt.Println("func JWTAuthzMiddleware.keyUserID", userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
