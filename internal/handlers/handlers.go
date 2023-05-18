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

	link := req.FormValue("link")
	struna := string(RandStringRunes(6))

	if contains(struna) {
		struna = string(RandStringRunes(6))
		contains(struna)
	} else {
		fullink := linkey[struna]
		if fullink == "" {
			linkey[struna] = link
			res.WriteHeader(http.StatusCreated)
			body := fmt.Sprintf("Ваша ссылка готова - localhost:8080/%s", struna)
			res.Write([]byte(body))
		}
	}

}

func GetPage(res http.ResponseWriter, req *http.Request) {
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
}

func contains(struna string) bool {
	for _, element := range linkey {
		if element == struna {
			return true
		}
	}
	return false
}
