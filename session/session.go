package session

import "time"

// Session is the data structure that holds the information we want to remember,
// plus some metadata. This is only stored server-side, the client will only get
// the ID.
type Session struct {
	id               string
	csrfToken        string
	username         string
	expires          time.Time
	oauthState       string
	oauthProvider    string
	oauthDestination string
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Username() string {
	return s.username
}

func (s *Session) SetUsername(username string) {
	s.username = username
}

func (s *Session) CsrfToken() string {
	return s.csrfToken
}

func (s *Session) OAuthState() string {
	return s.oauthState
}

func (s *Session) SetOAuthState(oauthState string) {
	s.oauthState = oauthState
}

func (s *Session) OAuthProvider() string {
	return s.oauthProvider
}

func (s *Session) SetOAuthProvider(oauthProvider string) {
	s.oauthProvider = oauthProvider
}

func (s *Session) OAuthDestination() string {
	return s.oauthDestination
}

func (s *Session) SetOAuthDestination(oauthDestination string) {
	s.oauthDestination = oauthDestination
}

func (s *Session) Touch(ttl time.Duration) {
	s.expires = time.Now().Add(ttl)
}

func (s *Session) End() {
	s.expires = time.Unix(0, 0)
}

func (s *Session) isExpired() bool {
	return s.expires.Before(time.Now())
}
