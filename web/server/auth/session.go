package server

import (
	"github.com/gorilla/sessions"
	"net/http"
	"yeetfile/utils"
)

var (
	key   = []byte(utils.GenRandomString(16))
	store = sessions.NewCookieStore(key)
)

func GetSession(req *http.Request) (*sessions.Session, error) {
	return store.Get(req, "session")
}
