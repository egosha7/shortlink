package routes

import (
	"github.com/egosha7/shortlink/internal/compress"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/storage"
	"net/http"

	"github.com/go-chi/chi"
)

func SetupRoutes(cfg *config.Config, store *storage.URLStore) http.Handler {

	// Создание роутера
	r := chi.NewRouter()
	gzipMiddleware := compress.GzipMiddleware{}

	r.Group(
		func(route chi.Router) {
			route.Use(gzipMiddleware.Apply)
			route.Get(
				"/{id}", func(w http.ResponseWriter, r *http.Request) {
					handlers.RedirectURL(w, r, store)
				},
			)

			route.Post(
				"/", func(w http.ResponseWriter, r *http.Request) {
					handlers.ShortenURL(w, r, cfg, store)
				},
			)

			route.Post(
				"/api/shorten", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenURL(w, r, cfg, store)
				},
			)
		},
	)

	return r
}
