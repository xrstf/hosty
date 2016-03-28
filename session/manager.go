package session

import (
	"net/http"
	"sync"
	"time"
)

// manager is the global manager of all sessions, holding a map of each
// issues session and regularly cleaning up expired sessions.
type manager struct {
	lock     sync.RWMutex
	sessions map[string]*Session
	idBytes  int
	ttl      time.Duration
	cookie   CookieOptions
}

func (sm *manager) NewContext(request *http.Request, response http.ResponseWriter) Context {
	var s *Session

	// find the session cookie
	cookie, err := request.Cookie(sm.cookie.Name)
	if err == nil {
		// check if this session exists
		s = sm.Get(cookie.Value)
		if s != nil {
			// if it already expired but got not cleaned up yet, ignore it;
			// otherwise, update the expiry time.
			if s.isExpired() {
				sm.Destroy(s.id)
				s = nil
			}
		}
	}

	ctx := Context{response, sm, s}
	ctx.Touch() // update expiry and set a precursory new Cookie header to update the client's expiry

	return ctx
}

func (sm *manager) NewSession() (*Session, error) {
	id, err := RandomString(sm.idBytes)
	if err != nil {
		return nil, err
	}

	csrfToken, err := RandomString(sm.idBytes)
	if err != nil {
		return nil, err
	}

	s := &Session{
		id:        id,
		csrfToken: csrfToken,
	}
	s.Touch(sm.ttl)

	sm.lock.Lock()
	sm.sessions[id] = s
	sm.lock.Unlock()

	return s, nil
}

func (sm *manager) Get(id string) *Session {
	sm.lock.RLock()
	s, _ := sm.sessions[id]
	sm.lock.RUnlock()

	return s
}

func (sm *manager) Destroy(id string) {
	sm.lock.Lock()
	delete(sm.sessions, id)
	sm.lock.Unlock()
}

func (sm *manager) Cleanup() {
	sm.lock.Lock()

	for id, s := range sm.sessions {
		if s.isExpired() {
			delete(sm.sessions, id)
		}
	}

	sm.lock.Unlock()
}
