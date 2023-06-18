package helpers

import (
	"math/rand"
)

func GenerateID(n int) string {
	// генерация случайного идентификатора
	runestring := "34" // abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
	var letterRunes = []rune(runestring)
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
