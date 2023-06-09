package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/services"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func ShortenURL(cfg *config.Config, store *storage.URLStore) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var req ShortenURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			id := services.GenerateID(6)
			url, ok := store.GetURL(id)
			if !ok {
				store.AddURL(id, req.URL)
			} else {
				fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
			}

			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, shortURL)
		},
	)
}

func HandleShortenURL(cfg *config.Config, store *storage.URLStore) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var req ShortenURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			id := services.GenerateID(6)
			url, ok := store.GetURL(id)
			if !ok {
				store.AddURL(id, req.URL)
			} else {
				fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
			}

			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			response := struct {
				Result string `json:"result"`
			}{
				Result: shortURL,
			}

			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		},
	)
}

func RedirectURL(store *storage.URLStore) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			url, ok := store.GetURL(id)
			if !ok {
				http.Error(w, "Invalid URL", http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		},
	)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок Accept-Encoding для поддержки сжатия gzip
			encodings := r.Header.Get("Accept-Encoding")
			if strings.Contains(encodings, "gzip") {
				// Устанавливаем заголовок Content-Encoding для указания сжатия gzip
				w.Header().Set("Content-Encoding", "gzip")

				// Создаем новый gzipResponseWriter
				gzipWriter := gzip.NewWriter(w)
				defer gzipWriter.Close()

				// Обновляем ResponseWriter для использования gzipResponseWriter
				w = &gzipResponseWriter{
					ResponseWriter: w,
					gzipWriter:     gzipWriter,
				}
			}

			// Вызываем следующий обработчик
			next.ServeHTTP(w, r)
		},
	)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (grw *gzipResponseWriter) Write(data []byte) (int, error) {
	return grw.gzipWriter.Write(data)
}

func (grw *gzipResponseWriter) Flush() {
	grw.gzipWriter.Flush()
	grw.ResponseWriter.(http.Flusher).Flush()
}

func (grw *gzipResponseWriter) WriteHeader(statusCode int) {
	grw.ResponseWriter.WriteHeader(statusCode)
}
