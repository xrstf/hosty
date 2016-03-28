package oauth

import (
	"net/http"

	"github.com/xrstf/hosty/session"
)

type Provider interface {
	Start(sess *session.Session) (string, error)
	Finish(sess *session.Session, request *http.Request) (string, error)
	UserProfile(accessToken string) (string, string, error)
}
