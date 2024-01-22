package session

import (
	"github.com/gorilla/sessions"
	"net/http"
	"yeetfile/web/utils"
)

var (
	key = utils.GetEnvVar(
		"YEETFILE_SESSION_KEY",
		utils.GenRandomString(16))
	store = sessions.NewFilesystemStore("", []byte(key))
)

const ValueKey = "auth"
const SessionKey = "session"
const UserKey = "user"

func GetSession(req *http.Request) (*sessions.Session, error) {
	return store.Get(req, SessionKey)
}

func SetSession(id string, w http.ResponseWriter, req *http.Request) error {
	session, _ := GetSession(req)
	session.Values[ValueKey] = true
	session.Values[UserKey] = id
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
	session.Values[UserKey] = ""
	return session.Save(req, w)
}

func GetSessionUserID(session *sessions.Session) string {
	return session.Values[UserKey].(string)
}
