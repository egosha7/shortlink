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
	"os"
	"strings"
	"sync"
)

type URL struct {
	ID  string
	URL string
}

type URLStore struct {
	urls     []URL
	mu       sync.RWMutex
	filePath string
}

func NewURLStore(filePath string) *URLStore {
	return &URLStore{
		urls:     make([]URL, 0),
		filePath: filePath,
	}
}

func (s *URLStore) AddURL(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newURL := URL{ID: id, URL: url}
	s.urls = append(s.urls, newURL)

	// Сохранение данных в файл
	err := s.SaveToFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving data to file: %v\n", err)
	}
}

func (s *URLStore) GetURL(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.urls {
		if u.ID == id {
			return u.URL, true
		}
	}
	return "", false
}

func (s *URLStore) LoadFromFile() error {

	// Проверка существования файла
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Создание нового файла
		file, err := os.Create(s.filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Запись начальных данных в файл
		initialData := []byte("[{\"ID\":\"def456\",\"URL\":\"https://google.com\"}]")
		_, err = file.Write(initialData)
		if err != nil {
			return err
		}
	}

	// Открываем файл
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&s.urls)
	if err != nil {
		return err
	}

	return nil
}

func (s *URLStore) SaveToFile() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(s.urls)
	if err != nil {
		return err
	}

	return nil
}

func ShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *URLStore) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else if r.Method == "POST" {
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

			store.AddURL(id, string(body))
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

			store.AddURL(id, string(body))
			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, shortURL)
		}

		// Продолжение кода для обработки запроса
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

	// Проверяем заголовок Content-Encoding на наличие сжатия gzip
	if r.Header.Get("Content-Encoding") == "gzip" {
		// Читаем тело запроса с помощью gzip.NewReader
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to read request body: %w", err)
		}
		defer gzipReader.Close()

		// Читаем распакованное тело запроса
		body, err := io.ReadAll(gzipReader)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return "", fmt.Errorf("failed to read request body: %w", err)
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
