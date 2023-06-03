package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/services"
	"github.com/go-chi/chi"
	"io"
	"net/http"
	"sync"
)

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]string),
	}
}

func (s *URLStore) AddURL(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[id] = url
}

func (s *URLStore) GetURL(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[id]
	return url, ok
}

func ShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *URLStore) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else if r.Method == "POST" {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body) // Заменено на io.ReadAll
		if err != nil {
			http.Error(w, " not allowed", http.StatusBadGateway)
			return
		}

		id := services.GenerateID(6)

		url, ok := store.GetURL(id)
		if !ok {
			store.AddURL(id, string(body))
		} else {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес: ", url)
		}

		store.AddURL(id, string(body))
		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	}
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *URLStore) (string, error) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return "", fmt.Errorf("method not allowed")
	}

	defer r.Body.Close()

	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return "", fmt.Errorf("failed to decode request body: %w", err)
	}

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

func RedirectURL(w http.ResponseWriter, r *http.Request, store *URLStore) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else if r.Method == "GET" {
		id := chi.URLParam(r, "id")
		url, ok := store.GetURL(id)
		if !ok {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// NewGzipResponseWriter создает новый gzip.Writer поверх ResponseWriter
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

// Close закрывает gzip.Writer и завершает запись
func (grw *gzipResponseWriter) Close() error {
	grw.gzipWriter.Close()
	return nil
}
