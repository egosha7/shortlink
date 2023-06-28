package cookieMW

import (
	"net/http"

	"github.com/egosha7/shortlink/internal/handlers"
)

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Проверяем наличие и действительность куки
			_, err := r.Cookie(handlers.CookieName)
			if err != nil {
				handlers.SetCookieHandler(w, r)
			} else {
				// Куки действительна, продолжаем выполнение следующего обработчика
				next.ServeHTTP(w, r)
			}
		},
	)
}
