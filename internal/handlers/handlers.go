package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/services"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"io"
	"net/http"
)

func ShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) {

	// Проверяем заголовок Content-Encoding на наличие сжатия gzip
	if r.Header.Get("Content-Encoding") == "gzip" {
		// Читаем тело запроса с помощью gzip.NewReader
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer gzipReader.Close()

		// Читаем распакованное тело запроса
		body, err := io.ReadAll(gzipReader)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Используем распакованное тело запроса
		id := services.GenerateID(6)

		url, ok := store.GetURL(id)
		if !ok {
			store.AddURL(id, string(body))
		} else {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
		}
		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)

	} else {
		// Читаем тело запроса без сжатия
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Используем тело запроса
		id := services.GenerateID(6)

		url, ok := store.GetURL(id)
		if !ok {
			store.AddURL(id, string(body))
		} else {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
		}

		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	}
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) (string, error) {

	if r.Header.Get("Content-Encoding") == "gzip" {
		// Читаем тело запроса с помощью gzip.NewReader
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to read request body: %w", err)
		}
		defer gzipReader.Close()

		// Декодируем распакованное тело запроса
		var req ShortenURLRequest
		err = json.NewDecoder(gzipReader).Decode(&req)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to decode request body: %w", err)
		}

		// Используем распакованное тело запроса
		id := services.GenerateID(6)

		url, ok := store.GetURL(id)
		if !ok {
			store.AddURL(id, req.URL)
		} else {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
		}

		store.AddURL(id, req.URL)
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
			return "", fmt.Errorf("failed to encode response: %w", err)
		}

		return shortURL, nil

	} else {
		// Читаем тело запроса без сжатия
		var req ShortenURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to decode request body: %w", err)
		}

		// Используем тело запроса
		id := services.GenerateID(6)

		url, ok := store.GetURL(id)
		if !ok {
			store.AddURL(id, req.URL)
		} else {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
		}

		store.AddURL(id, req.URL)
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
			return "", fmt.Errorf("failed to encode response: %w", err)
		}

		return shortURL, nil
	}
}

func RedirectURL(w http.ResponseWriter, r *http.Request, store *storage.URLStore) {

	id := chi.URLParam(r, "id")
	url, ok := store.GetURL(id)
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок Content-Encoding на наличие сжатия gzip
			if r.Header.Get("Content-Encoding") == "gzip" {
				// Читаем тело запроса с помощью gzip.NewReader
				gzipReader, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}
				defer gzipReader.Close()

				// Читаем распакованное тело запроса
				body, err := io.ReadAll(gzipReader)
				if err != nil {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}

				// Передаем распакованное тело запроса в следующий обработчик
				r.Body = io.NopCloser(bytes.NewReader(body))

				// Вызываем следующий обработчик в цепочке
				next.ServeHTTP(w, r)

			} else {
				// Вызываем следующий обработчик в цепочке
				next.ServeHTTP(w, r)
			}
		},
	)
}

// NewGzipResponseWriter создает новый gzipResponseWriter поверх ResponseWriter
func NewGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
		gzipWriter:     gzip.NewWriter(w),
	}
}

// gzipResponseWriter оборачивает http.ResponseWriter для поддержки сжатия gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write перехватывает запись данных и передает их в gzip.Writer
func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.gzipWriter.Write(b)
}

// Header возвращает заголовки ResponseWriter
func (grw *gzipResponseWriter) Header() http.Header {
	return grw.ResponseWriter.Header()
}

// WriteHeader записывает статусный код ответа
func (grw *gzipResponseWriter) WriteHeader(statusCode int) {
	grw.ResponseWriter.Header().Del("Content-Length")
	grw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	grw.ResponseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и завершает запись
func (grw *gzipResponseWriter) Close() error {
	grw.gzipWriter.Close()
	return nil
}
