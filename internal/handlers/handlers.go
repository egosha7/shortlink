package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/services"
	"github.com/go-chi/chi"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
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
	}

	// Читаем тело запроса в []byte
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *URLStore) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса в структуру ShortenURLRequest
	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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
		return
	}
}

func RedirectURL(w http.ResponseWriter, r *http.Request, store *URLStore) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

			// Вызываем следующий обработчик в цепочке
			next.ServeHTTP(w, r)
		},
	)
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

// WriteHeader записывает заголовки и код состояния ответа
func (grw *gzipResponseWriter) WriteHeader(code int) {
	// Если клиент не поддерживает сжатие gzip, отправляем несжатый ответ
	if code != http.StatusNoContent && code != http.StatusNotModified && code >= 200 && grw.Header().Get("Content-Encoding") != "gzip" {
		grw.ResponseWriter.Header().Del("Content-Encoding")
		grw.gzipWriter.Reset(ioutil.Discard)
		grw.ResponseWriter.WriteHeader(code)
	} else {
		grw.ResponseWriter.WriteHeader(code)
	}
}

// Header возвращает заголовки ResponseWriter
func (grw *gzipResponseWriter) Header() http.Header {
	return grw.ResponseWriter.Header()
}
