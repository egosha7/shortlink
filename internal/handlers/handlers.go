package handlers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/services"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"io"
	"net/http"
	"strings"
)

type key int

const uncompressedBodyKey key = iota

func ShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) {

	var body []byte
	var err error

	// Проверяем, есть ли распакованное тело запроса в контексте
	if uncompressedBody, ok := r.Context().Value(uncompressedBodyKey).(*gzip.Reader); ok {
		// Если есть распакованное тело, читаем данные из него
		body, err = io.ReadAll(uncompressedBody)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Если распакованного тела нет, читаем данные из исходного тела запроса
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

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

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) (string, error) {

	var body []byte
	var err error

	// Проверяем, есть ли распакованное тело запроса в контексте
	if uncompressedBody, ok := r.Context().Value(uncompressedBodyKey).(*gzip.Reader); ok {
		// Если есть распакованное тело, читаем данные из него
		body, err = io.ReadAll(uncompressedBody)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return "", fmt.Errorf("failed to decode request body: %w", err)
		}
	} else {
		// Если распакованного тела нет, читаем данные из исходного тела запроса
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to decode request body: %w", err)
		}
	}

	// Декодируем распакованное тело запроса
	var req ShortenURLRequest
	err = json.Unmarshal(body, &req)
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
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				// Если запрос поддерживает сжатие gzip, пропускаем его дальше
				next.ServeHTTP(w, r)
			} else {
				// Если запрос не поддерживает сжатие gzip, создаем новый Request с распакованным телом
				uncompressedBody, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				defer uncompressedBody.Close()

				// Создаем новый Request с распакованным телом
				uncompressedRequest := r.WithContext(
					context.WithValue(
						r.Context(), uncompressedBodyKey, uncompressedBody,
					),
				)

				// Пропускаем новый Request дальше
				next.ServeHTTP(w, uncompressedRequest)
			}
		},
	)
}

// NewGzipResponseWriter создает новый gzipResponseWriter, оборачивающий http.ResponseWriter и gzip.Writer
func NewGzipResponseWriter(w http.ResponseWriter, gzipWriter *gzip.Writer) http.ResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
		gzipWriter:     gzipWriter,
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
