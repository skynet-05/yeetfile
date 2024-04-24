package session

import (
	"errors"
	"github.com/gorilla/sessions"
	"net/http"
	"yeetfile/shared"
	"yeetfile/web/utils"
)

type HandlerFunc func(w http.ResponseWriter, req *http.Request, userID string)

var (
	key = utils.GetEnvVar(
		"YEETFILE_SESSION_KEY",
		shared.GenRandomString(16))
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
	session, err := GetSession(req)
	if err != nil {
		return false
	}

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
	sessionVal := session.Values[UserKey]
	if sessionVal != nil {
		return sessionVal.(string)
	}

	return ""
}

func GetSessionAndUserID(req *http.Request) (string, error) {
	s, err := GetSession(req)
	if err != nil {
		return "", errors.New("invalid session")
	}

	id := GetSessionUserID(s)
	if len(id) == 0 {
		return "", errors.New("invalid session")
	}

	return id, nil
}
