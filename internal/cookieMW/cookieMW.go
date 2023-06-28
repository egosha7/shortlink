package cookieMW

import (
	"github.com/egosha7/shortlink/internal/handlers"
	"net/http"
)

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Получаем значение куки

			_, err := r.Cookie(handlers.CookieName)
			if err != nil {
				handlers.SetCookieHandler(w, r)
			}

			// Продолжаем выполнение следующего обработчика
			next.ServeHTTP(w, r)
		},
	)
}
