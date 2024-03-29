package auth

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/helpers"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const CookieName = "USER_ID"

// Функция для генерации симметрично подписанной куки с помощью JWT
func SetSignedCookie(w http.ResponseWriter, userID string, secretKey []byte, expiration time.Duration) {
	// Создаем новый токен
	token := jwt.New(jwt.SigningMethodHS256)

	// Устанавливаем идентификатор подписчика (subject) и значение userID в токене
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID

	// Устанавливаем срок действия токена
	expirationTime := time.Now().Add(expiration)
	claims["exp"] = expirationTime.Unix()

	// Подписываем токен с использованием секретного ключа
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Создаем новую куку с подписанным токеном
	cookie := http.Cookie{
		Name:     CookieName,
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
	}

	// Устанавливаем куку в ответе
	http.SetCookie(w, &cookie)
}

// Функция для проверки симметрично подписанной куки с помощью JWT
func VerifySignedCookie(r *http.Request, secretKey []byte) (string, error) {
	// Получаем значение куки из запроса
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", fmt.Errorf("сookie not found")
	}

	// Проверяем подпись куки и получаем токен
	token, err := jwt.Parse(
		cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что используется правильный алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			// Возвращаем секретный ключ для проверки подписи
			return secretKey, nil
		},
	)

	if err != nil {
		return "", err
	}

	// Проверяем, что токен действителен и получаем значение userID из токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", fmt.Errorf("invalid cookie")
}

func SetCookieHandler(w http.ResponseWriter, r *http.Request) string {

	// Генерируем уникальный идентификатор пользователя
	userID := helpers.GenerateID(8)

	// Получаем секретный ключ для подписи
	secretKey := []byte("your-secret-key")

	// Устанавливаем куку
	SetSignedCookie(w, userID, secretKey, time.Hour*24)

	// Отправляем ответ
	return userID
}

func GetCookieHandler(w http.ResponseWriter, r *http.Request) string {

	// Получаем секретный ключ для подписи
	secretKey := []byte("your-secret-key")

	// Проверяем подпись и извлекаем userID
	userID, err := VerifySignedCookie(r, secretKey)
	if err != nil {
		return ""
	}

	return userID
}
