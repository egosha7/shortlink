package cookieMW

import (
	"context"
	"github.com/egosha7/shortlink/internal/handlers"
	"net/http"
)

type UserData struct {
	UserID string
}

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			var userData UserData
			// Получаем значение куки

			cookie, err := r.Cookie(handlers.CookieName)
			if err != nil || cookie == nil {
				// Кука не существует
				id := handlers.SetCookieHandler(w, r)
				userData.UserID = id
				ctx := context.WithValue(r.Context(), "userID", id)
				r = r.WithContext(ctx)
			}

			// Продолжаем выполнение следующего обработчика
			next.ServeHTTP(w, r)
		},
	)
}
