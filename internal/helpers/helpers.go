package helpers

import (
	"github.com/google/uuid"
)

var LastGeneratedID string

func GenerateID(n int) string {
	id := uuid.New().String()
	if n < len(id) {
		LastGeneratedID = id[:n]
		return id[:n]
	}
	LastGeneratedID = id
	return id
}
