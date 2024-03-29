package cookiemw

import (
	"context"
	"github.com/egosha7/shortlink/internal/auth"
	"github.com/egosha7/shortlink/internal/handlers"
	"net/http"
)

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Получаем значение куки

			cookie, err := r.Cookie(auth.CookieName)
			if err != nil || cookie == nil {
				// Кука не существует
				id := auth.SetCookieHandler(w, r)
				ctx := context.WithValue(r.Context(), handlers.UserIDKey, id)
				r = r.WithContext(ctx)
			}

			// Продолжаем выполнение следующего обработчика
			next.ServeHTTP(w, r)
		},
	)
}
