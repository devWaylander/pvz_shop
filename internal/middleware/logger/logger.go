package logger

import (
	"net/http"

	"github.com/devWaylander/coins_store/pkg/log"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Logger.Info().Msgf("%s --- %s ---", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
