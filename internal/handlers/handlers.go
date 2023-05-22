package handlers

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/common/config"
	"github.com/egosha7/shortlink/internal/otherfunc"
	"io"
	"net/http"
)

var urls = make(map[string]string)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body) // Заменено на io.ReadAll
	if err != nil {
		http.Error(w, "Method not allowed", http.StatusBadGateway)
		return
	}

	id := otherfunc.GenerateID(6)
	cfg := config.New()

	urls[id] = string(body)
	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL) // Заменено на fmt.Fprint
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[1:]
	url, ok := urls[id]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
