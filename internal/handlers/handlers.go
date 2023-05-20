package handlers

import (
	"egosha7/shortlink/internal/OtherFunc"
	"fmt"
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

	urls[id] = string(body)
	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, shortURL)
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
