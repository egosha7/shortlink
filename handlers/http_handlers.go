// handlers/http_handlers.go

package handlers

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/internal/auth"
	"go.uber.org/zap"
	"golang.org/x/tools/refactor/rename"
	"io"
	"net/http"

	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/logic"
)

// ContextKey представляет ключ контекста для идентификатора пользователя.
type ContextKey string

// UserIDKey - ключ контекста, используемый для хранения идентификатора пользователя.
const UserIDKey ContextKey = "userID"

// ShortenURL сокращает URL и возвращает короткую ссылку.
func ShortenURL(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore, logger *zap.Logger) {
	userID := auth.GetCookieHandler(w, r)

	setCookieHeader := w.Header().Get("Set-Cookie")
	if setCookieHeader != "" {
		userID = r.Context().Value(UserIDKey).(string)
		newCtx := context.WithValue(r.Context(), UserIDKey, "")
		r = r.WithContext(newCtx)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := logic.ShortenURL(body, userID, store, BaseURL)
	if err != nil && err == rename.ConflictError {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURL))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}
