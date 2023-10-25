package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/auth"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type Key string

type ContextKey string

const UserIDKey ContextKey = "userID"

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

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Отправляем ссылки и userID в канал через метод DeleteURLs экземпляра worker
	wkr.DeleteURLs(urls, userID)

	// Возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)
}

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

	// Получение сокращенных URL пользователя из хранилища
	urls := store.GetURLsByUserID(userID)

	if len(urls) == 0 {
		// Если нет сокращенных URL пользователя, возвращаем статус 204 No Content
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Формируем ответ в формате JSON
	var response []map[string]string
	for _, u := range urls {
		response = append(
			response, map[string]string{
				"short_url":    BaseURL + "/" + u.ID,
				"original_url": u.URL,
			},
		)
	}

	// Отправка ответа в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func ShortenURL(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore, logger *zap.Logger) {
	id := helpers.GenerateID(6)

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

	logger.Info("Request body (POST /)" + string(body))

	var existingID string
	var switchBool bool
	existingID, switchBool = store.AddURL(id, string(body), userID)
	if existingID != "" && !switchBool {
		existingID = strings.TrimRight(existingID, "\n")
		shortURLout := fmt.Sprintf("%s/%s", BaseURL, existingID)
		fmt.Println("По этому адресу уже зарегистрирован другой адрес:")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURLout))
		return
	} else if existingID != "" && switchBool {
		id = existingID
	}

	shortURL := fmt.Sprintf("%s/%s", BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) (string, error) {

	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return "", fmt.Errorf("failed to decode request body: %w", err)
	}

	userID := auth.GetCookieHandler(w, r)

	// Используем тело запроса
	id := helpers.GenerateID(6)

	var existingID string
	var switchBool bool
	existingID, switchBool = store.AddURL(id, req.URL, userID)
	if existingID != "" && !switchBool {
		fmt.Println("По этому адресу уже зарегистрирован другой адрес:", existingID)

		response := struct {
			Result string `json:"result"`
		}{
			Result: existingID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return "", fmt.Errorf("failed to encode response: %w", err)
		}
		return "", fmt.Errorf("failed to save URL to database")
	} else if existingID != "" && switchBool {
		id = existingID
	}

	shortURL := fmt.Sprintf("%s/%s", BaseURL, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", fmt.Errorf("failed to encode response: %w", err)
	}

	return shortURL, nil
}

func RedirectURL(w http.ResponseWriter, r *http.Request, store *storage.URLStore) {
	id := chi.URLParam(r, "id")
	url, ok := store.GetURL(id)
	if url == "" && !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	} else if !ok {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleShortenBatch(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) {

	ctx := r.Context()

	userID := auth.GetCookieHandler(w, r)

	var records []map[string]string
	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверяем, что есть записи для обработки
	if len(records) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	res, _ := store.AddURLwithTx(records, ctx, BaseURL, userID)
	if res == nil {
		http.Error(w, "StatusBadRequest", http.StatusBadRequest)
		return
	}
	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
