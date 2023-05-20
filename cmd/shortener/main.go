package main

import (
	"egosha7/shortlink/internal/handlers"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.ShortenURL)
	http.HandleFunc("/{id}", handlers.RedirectURL)
	http.ListenAndServe(":8080", nil)
}
