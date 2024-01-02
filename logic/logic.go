// logic/logic.go

package logic

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/storage"
	"golang.org/x/tools/refactor/rename"
	"strings"
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
