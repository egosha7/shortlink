package compress

import (
	"compress/gzip"
	"net/http"
)

// GzipMiddleware - это структура для GzipMiddleware.
type GzipMiddleware struct{}

// Apply - это метод, который применяет GzipMiddleware к следующему обработчику HTTP.
func (m *GzipMiddleware) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, содержит ли заголовок `Content-Encoding` значение `gzip`.
			if r.Header.Get(`Content-Encoding`) == `gzip` {
				// Если это так, создаем новый Gzip Reader для тела запроса.
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					// В случае ошибки отправляем HTTP ошибку и выходим.
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer gz.Close()

				// Замена тела запроса на распакованное содержимое с ограничением по размеру.
				r.Body = http.MaxBytesReader(w, gz, r.ContentLength)
				r.Header.Del("Content-Encoding")
				r.Header.Del("Content-Length")
			}

			// Передаем управление следующему обработчику в цепочке.
			next.ServeHTTP(w, r)
		},
	)
}
