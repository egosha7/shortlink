package routes

import (
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/storage"
	"net/http"

	"github.com/go-chi/chi"
)

func SetupRoutes(cfg *config.Config, store *storage.URLStore) http.Handler {

	// Создание роутера
	r := chi.NewRouter()
	r.Use(handlers.GzipMiddleware)
	r.Method(http.MethodGet, "/{id}", handlers.RedirectURL(store))
	r.Method(http.MethodPost, "/", handlers.ShortenURL(cfg, store))
	r.Method(http.MethodPost, "/api/shorten", handlers.HandleShortenURL(cfg, store))

	return r
}
