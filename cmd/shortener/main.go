package main

import (
	"egosha7/shortlink/internal/handlers"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"os"
)

func main() {
	r := chi.NewRouter()
	r.Get(`/`, handlers.RedirectURL)
	r.Post(`/`, handlers.ShortenURL)
	r.NotFound(handlers.RedirectURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
