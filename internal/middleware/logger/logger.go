package logger

import (
	"net/http"

	"github.com/devWaylander/pvz_store/pkg/log"
)

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Logger.Info().Msgf("%s --- %s ---", r.Method, r.URL.Path)
			next.ServeHTTP(w, r) // Передаем запрос следующему обработчику
		})
	}
}
