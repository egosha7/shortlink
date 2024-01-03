package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/auth"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/egosha7/shortlink/logic"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// ContextKey представляет ключ контекста для идентификатора пользователя.
type ContextKey string

// UserIDKey - ключ контекста, используемый для хранения идентификатора пользователя.
const UserIDKey ContextKey = "userID"

// ShortenURLRequest представляет структуру запроса для сокращения URL.
type ShortenURLRequest struct {
	URL string `json:"url"`
}

// DeleteUserURLsHandler удаляет URL'ы, принадлежащие пользователю.
func DeleteUserURLsHandler(w http.ResponseWriter, r *http.Request, wkr *worker.Worker) {
	userID := auth.GetCookieHandler(w, r)
	setCookieHeader := w.Header().Get("Set-Cookie")
	if setCookieHeader != "" {
		fmt.Println("Cookie set in the response:", setCookieHeader)
		userID = r.Context().Value(UserIDKey).(string)
		newCtx := context.WithValue(r.Context(), UserIDKey, "")
		r = r.WithContext(newCtx)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logic.DeleteUserURLs(body, userID, wkr)

	// Возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)
}

// GetUserURLsHandler возвращает URL'ы, принадлежащие пользователю.
func GetUserURLsHandler(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore, logger *zap.Logger) {
	// Получение идентификатора пользователя из куки
	userID := auth.GetCookieHandler(w, r)
	setCookieHeader := w.Header().Get("Set-Cookie")
	if setCookieHeader != "" {
		logger.Info("Cookie set in the response:" + setCookieHeader)
		userID = r.Context().Value(UserIDKey).(string)
		newCtx := context.WithValue(r.Context(), UserIDKey, "")
		r = r.WithContext(newCtx)
	} else if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := logic.GetUserURLs(BaseURL, userID, store)
	if err != nil {
		// Обработка ошибки
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ в формате JSON
	var jsonResponse []map[string]string
	for _, u := range response {
		jsonResponse = append(
			jsonResponse, map[string]string{
				"short_url":    u["short_url"],
				"original_url": u["original_url"],
			},
		)
	}

	// Отправка ответа в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}

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
	if err != nil {
		// Обработка ошибки
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

// HandleShortenURL обрабатывает запрос на сокращение URL.
func HandleShortenURL(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := auth.GetCookieHandler(w, r)

	shortURL, err := logic.HandleShortenURL(body, userID, store, BaseURL)
	if err != nil {
		// Обработка ошибки
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

// HandleShortenBatch обрабатывает пакетный запрос на сокращение URL.
func HandleShortenBatch(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore, logger *zap.Logger) {

	ctx := r.Context()

	userID := auth.GetCookieHandler(w, r)

	var records []map[string]string
	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		logger.Error("1", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	res, _ := logic.HandleShortenBatch(records, ctx, BaseURL, store, userID)
	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// StatsHandler возвращает статистику о сокращенных URL и пользователях.
func StatsHandler(w http.ResponseWriter, r *http.Request, store *storage.URLStore, trustedSubnet string) {
	// Получение IP-адреса клиента из заголовка X-Real-IP
	clientIP := r.Header.Get("X-Real-IP")
	// Проверка, что IP-адрес клиента находится в доверенной подсети.
	if !logic.IsTrustedClient(clientIP, trustedSubnet) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	stats, _ := logic.StatsHandler(store)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
