package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"io"
	"net/http"
	"strings"
)

type Key string

func GetUserURLsHandler(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) {
	// Получение идентификатора пользователя из куки
	userID := GetCookieHandler(w, r)

	_, err := r.Cookie(CookieName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получение сокращенных URL пользователя из хранилища
	urls := store.GetURLsByUserID(userID)

	if len(urls) == 0 {
		// Если нет сокращенных URL пользователя, возвращаем статус 204 No Content
		w.WriteHeader(http.StatusNoContent)
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

func ShortenURL(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) {
	id := helpers.GenerateID(6)

	userID := GetCookieHandler(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	userID := GetCookieHandler(w, r)

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
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleShortenBatch(w http.ResponseWriter, r *http.Request, BaseURL string, store *storage.URLStore) {

	ctx := r.Context()

	userID := GetCookieHandler(w, r)

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
