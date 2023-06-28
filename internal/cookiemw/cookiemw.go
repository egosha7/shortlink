package cookiemw

import (
	"context"
	"github.com/egosha7/shortlink/internal/handlers"
	"net/http"
)

type contextKey string

const UserIDKey = contextKey("userID")

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(handlers.CookieName)
			if err != nil || cookie == nil {
				// Кука не существует
				id := handlers.SetCookieHandler(w, r)
				ctx := context.WithValue(r.Context(), UserIDKey, id)
				r = r.WithContext(ctx)
			}

			// Продолжаем выполнение следующего обработчика
			next.ServeHTTP(w, r)
		},
	)
}
