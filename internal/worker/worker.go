package worker

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/storage"
)

// Воркер
type Worker struct {
	worker chan string
	store  *storage.URLStore
	userID string
}

// Метод для добавления ссылки на удаление в канал
func (w *Worker) DeleteURLs(urls []string, userID string) {
	for _, url := range urls {
		w.worker <- url
	}
	w.userID = userID
}

// Метод для обработки ссылок на удаление
func (w *Worker) work() {
	for url := range w.worker {
		// Используйте хранилище для удаления ссылки
		err := w.store.DeleteURLs(url, w.userID)
		if err != nil {
			fmt.Println("Error deleting URL:", err)
		}
	}
}
