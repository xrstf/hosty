package session

import "net/http"

// Context is the context that is injected in the gin.Context and allows controllers
// to access and control the session (i.e. start and end sessions). This is the glue
// between the HTTP layer and the session logic, i.e. cookies are managed here.
type Context struct {
	http     http.ResponseWriter
	sessions *manager
	session  *Session
}

func (ctx *Context) Session() *Session {
	return ctx.session
}

func (ctx *Context) Username() string {
	if ctx.IsAuthenticated() {
		return ctx.session.Username()
	}

	return ""
}

func (ctx *Context) IsAuthenticated() bool {
	return ctx.session != nil && ctx.session.Username() != ""
}

func (ctx *Context) Start() (*Session, error) {
	// if there is already a session, kill it
	if ctx.session != nil {
		ctx.End()
	}

	session, err := ctx.sessions.NewSession()
	if err != nil {
		return session, err
	}

	ctx.session = session
	ctx.setCookie()

	return session, err
}

func (ctx *Context) End() {
	if ctx.session != nil {
		ctx.session.End()
		ctx.sessions.Destroy(ctx.session.id)
		ctx.unsetCookie()
		ctx.session = nil
	}
}

func (ctx *Context) Touch() {
	if ctx.session != nil {
		ctx.session.Touch(ctx.sessions.ttl)
		ctx.setCookie()
	}
}

func (ctx *Context) setCookie() {
	ctx.sendCookie(int(ctx.sessions.cookie.MaxAge.Seconds()))
}

func (ctx *Context) unsetCookie() {
	ctx.sendCookie(-1)
}

func (ctx *Context) sendCookie(maxAge int) {
	cookie := http.Cookie{
		Name:     ctx.sessions.cookie.Name,
		Value:    ctx.session.id,
		HttpOnly: true,
		MaxAge:   maxAge,
		Path:     ctx.sessions.cookie.Path,
		Secure:   ctx.sessions.cookie.Secure,
	}

	ctx.http.Header().Set("Set-Cookie", cookie.String())
}
