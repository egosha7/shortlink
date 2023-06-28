package helpers

import (
	"encoding/base64"
	"math/rand"
)

func GenerateID(n int) string {
	// генерация случайного идентификатора
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:n]
}
