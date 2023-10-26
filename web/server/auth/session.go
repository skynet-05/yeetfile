package auth

import (
	"github.com/gorilla/sessions"
	"net/http"
	"yeetfile/utils"
)

var (
	key = utils.GetEnvVar(
		"YEETFILE_SESSION_KEY",
		utils.GenRandomString(16))
	store = sessions.NewFilesystemStore(".", []byte(key))
)

const ValueKey = "auth"
const SessionKey = "session"

func GetSession(req *http.Request) (*sessions.Session, error) {
	return store.Get(req, SessionKey)
}

func SetSession(w http.ResponseWriter, req *http.Request) error {
	session, _ := GetSession(req)
	session.Values[ValueKey] = true
	return session.Save(req, w)
}

func IsValidSession(req *http.Request) bool {
	session, _ := GetSession(req)
	ok, found := session.Values[ValueKey].(bool)

	return found && ok
}

func RemoveSession(w http.ResponseWriter, req *http.Request) error {
	session, _ := GetSession(req)

	session.Values[ValueKey] = false
	return session.Save(req, w)
}
