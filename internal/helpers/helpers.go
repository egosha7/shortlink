package helpers

import (
	"github.com/google/uuid"
)

func GenerateID(n int) string {
	id := uuid.New().String()
	if n < len(id) {
		return id[:n]
	}
	return id
}
