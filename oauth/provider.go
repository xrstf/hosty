package oauth

import (
	"net/http"

	"go.xrstf.de/hosty/session"
)

type Provider interface {
	Start(sess *session.Session) (string, error)
	Finish(sess *session.Session, request *http.Request) (string, error)
	UserProfile(accessToken string) (string, string, error)
}
