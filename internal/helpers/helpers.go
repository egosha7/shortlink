package helpers

import (
	"github.com/google/uuid"
)

// LastGeneratedID хранит последний сгенерированный идентификатор.
var LastGeneratedID string

// GenerateID генерирует уникальный идентификатор длиной n с использованием UUID.
// Возвращает сгенерированный идентификатор.
func GenerateID(n int) string {
	id := uuid.New().String()
	if n < len(id) {
		// Если n меньше длины сгенерированного UUID, обрезаем UUID до n символов
		LastGeneratedID = id[:n]
		return id[:n]
	}
	// В противном случае, сохраняем полный UUID
	LastGeneratedID = id
	return id
}
