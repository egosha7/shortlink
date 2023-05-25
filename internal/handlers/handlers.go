package handlers

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/common/config"
	"github.com/egosha7/shortlink/internal/otherfunc"
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

		id := otherfunc.GenerateID(6)

		store.AddURL(id, string(body))
		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
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
