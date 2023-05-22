package handlers

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/OtherFunc"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
)

var urls = make(map[string]string)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// обработка ошибки
	}

	id := OtherFunc.GenerateID(6)
	cfg := config.New()

	urls[id] = string(body)
	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, shortURL)
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shortURL := chi.URLParam(r, "shortURL")
	url, ok := urls[shortURL]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	} else {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
