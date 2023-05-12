package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

type Config struct {
	Address string
	BaseURL string
}

const form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/" method="post">
            <label>Link </label><input type="text" name="link">
            <input type="submit" value="GO!">
        </form>
    </body>
</html>`

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var linkey = make(map[string]string)

func mainPage(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		if req.URL.Path == "/" {
			io.WriteString(res, form)
		} else {
			fullink := linkey[req.URL.Path[1:]]
			if fullink == "" {
				res.Write([]byte("404"))
			} else {
				http.Redirect(res, req, fullink, http.StatusSeeOther)
			}
			return
		}

	} else {
		link := req.FormValue("link")
		struna := string(RandStringRunes(6))

		linkey[struna] = link

		res.WriteHeader(http.StatusCreated)
		body := fmt.Sprintf("%s/%s", config.Address, struna)
		res.Write([]byte(body))
	}
}

var config Config

func init() {
	flag.StringVar(&config.Address, "a", ":8081", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()

	// инициализация полей из переменных окружения
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		config.Address = addr
	}
	if url := os.Getenv("BASE_URL"); url != "" {
		config.BaseURL = url
	}
}

func main() {
	r := chi.NewRouter()
	r.HandleFunc(`/`, mainPage)
	r.NotFound(mainPage)

	err := http.ListenAndServe(config.Address, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
	}
}
