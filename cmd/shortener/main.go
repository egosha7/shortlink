package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

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
				res.WriteHeader(http.StatusBadRequest)
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
		body := fmt.Sprintf(struna)
		res.Write([]byte(body))
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
