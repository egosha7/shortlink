// logic/logic.go

package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/tools/refactor/rename"
	"net"
	"net/http"
	"strings"

	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
)

// ShortenURL сокращает URL и возвращает короткую ссылку.
func ShortenURL(body []byte, userID string, store *storage.URLStore, BaseURL string) (string, error) {
	id := helpers.GenerateID(6)

	var existingID string
	existingID, switchBool := store.AddURL(id, string(body), userID)
	if existingID != "" && !switchBool {
		return fmt.Sprintf("%s/%s", BaseURL, strings.TrimRight(existingID, "\n")), error(rename.ConflictError)
	} else if existingID != "" && switchBool {
		return fmt.Sprintf("%s/%s", BaseURL, existingID), nil
	}

	return fmt.Sprintf("%s/%s", BaseURL, id), nil
}

func DeleteUserURLs(body []byte, userID string, wkr *worker.Worker) {
	var urls []string
	err := json.Unmarshal(body, &urls)
	if err != nil {
		// обработка ошибки, например, логгирование
		return
	}

	wkr.DeleteURLs(urls, userID)
}

// GetUserURLs возвращает URL'ы, принадлежащие пользователю.
func GetUserURLs(BaseURL string, userID string, store *storage.URLStore) ([]map[string]string, error) {
	urls := store.GetURLsByUserID(userID)

	if len(urls) == 0 {
		return nil, fmt.Errorf("unauthorized: %d", http.StatusUnauthorized)
	}

	var response []map[string]string
	for _, u := range urls {
		response = append(
			response, map[string]string{
				"short_url":    BaseURL + "/" + u.ID,
				"original_url": u.URL,
			},
		)
	}

	return response, nil
}

// HandleShortenURL обрабатывает запрос на сокращение URL.
func HandleShortenURL(body []byte, userID string, store *storage.URLStore, BaseURL string) (string, error) {
	id := helpers.GenerateID(6)

	var existingID string
	var switchBool bool
	existingID, switchBool = store.AddURL(id, string(body), userID)
	if existingID != "" && !switchBool {
		return fmt.Sprintf("%s/%s", BaseURL, strings.TrimRight(existingID, "\n")), error(rename.ConflictError)
	} else if existingID != "" && switchBool {
		id = existingID
	}

	return fmt.Sprintf("%s/%s", BaseURL, id), nil
}

// HandleShortenBatch обрабатывает пакетный запрос на сокращение URL.
func HandleShortenBatch(records []map[string]string, ctx context.Context, BaseURL string, store *storage.URLStore, userID string) ([]map[string]string, error) {
	// Проверяем, что есть записи для обработки
	if len(records) == 0 {
		return nil, fmt.Errorf("unauthorized: %d", http.StatusBadRequest)
	}

	res, _ := store.AddURLwithTx(records, ctx, BaseURL, userID)
	if res == nil {
		return nil, fmt.Errorf("unauthorized: %d", http.StatusBadRequest)
	}

	return res, nil
}

// HandleShortenBatch обрабатывает пакетный запрос на сокращение URL.
func StatsHandler(store *storage.URLStore) (map[string]int, error) {
	// Получение статистики
	urlsCount, usersCount := store.GetStats()

	// Отправка JSON-ответа
	stats := map[string]int{"urls": urlsCount, "users": usersCount}

	return stats, nil
}

// IsTrustedClient проверяет, что IP-адрес клиента находится в доверенной подсети.
func IsTrustedClient(clientIP string, trustedSubnet string) bool {
	// Парсинг подсети
	_, trustedIPNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		// Обработка ошибки, например, логгирование
		return false
	}

	// Парсинг IP-адреса клиента
	clientAddr := net.ParseIP(clientIP)
	if clientAddr == nil {
		// Обработка ошибки, например, логгирование
		return false
	}

	// Проверка, находится ли IP-адрес клиента в доверенной подсети
	return trustedIPNet.Contains(clientAddr)
}
