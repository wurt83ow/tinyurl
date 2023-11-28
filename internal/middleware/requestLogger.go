// Package middleware provides HTTP middleware logging functions.
package middleware

import (
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is an interface for logging operations.
type Log interface {
	Info(string, ...zapcore.Field)
}

// ReqLog is a middleware logger for incoming HTTP requests.
type ReqLog struct {
	log Log
}

// NewReqLog creates a new instance of ReqLog with the specified logger.
func NewReqLog(log Log) *ReqLog {
	return &ReqLog{
		log: log,
	}
}

// RequestLogger is an HTTP middleware that logs incoming requests.
func (rl *ReqLog) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.log.Info("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h.ServeHTTP(w, r)
	})
}
