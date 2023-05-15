package handlers

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/const"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune(_const.Runestring)

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var linkey = make(map[string]string)

func MainPage(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		if req.URL.Path == "/" {
			io.WriteString(res, _const.Form)
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
		body := fmt.Sprintf("%s", struna)
		res.Write([]byte(body))
	}
}
