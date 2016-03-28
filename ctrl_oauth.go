package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xrstf/hosty/oauth"
)

func setupOAuthCtrl(r *gin.Engine) {
	r.GET("/oauth/start", func(c *gin.Context) {
		ctx := getSessionContext(c)

		if ctx.IsAuthenticated() {
			c.Redirect(http.StatusFound, "/")
			return
		}

		// pure convenience: remember the destination throughout the OAuth flow
		dest := c.Query("destination")

		if len(dest) > 0 {
			_, err := store.FindByID(dest)
			if err != nil {
				dest = ""
			}
		}

		loginForm := func(status int, warning string) {
			renderLoginForm(c, status, dest, "", warning)
		}

		providerName := c.Query("provider")
		provider, err := buildOauthProvider(providerName)
		if err != nil {
			loginForm(http.StatusBadRequest, err.Error())
			return
		}

		sess, err := ctx.Start()
		if err != nil {
			loginForm(http.StatusInternalServerError, "Could not start a session: "+err.Error())
			return
		}

		sess.SetOAuthProvider(providerName)
		sess.SetOAuthDestination(dest)

		redirectURL, err := provider.Start(sess)
		if err != nil {
			loginForm(http.StatusInternalServerError, err.Error())
			return
		}

		c.Redirect(http.StatusFound, redirectURL)
	})

	r.GET("/oauth/callback", func(c *gin.Context) {
		sessionCtx := getSessionContext(c)
		sess := sessionCtx.Session()

		if sessionCtx.IsAuthenticated() || sess == nil {
			c.Redirect(http.StatusFound, "/")
			return
		}

		dest := sess.OAuthDestination()

		loginForm := func(status int, warning string) {
			renderLoginForm(c, status, dest, "", warning)
		}

		providerName := sess.OAuthProvider()
		provider, err := buildOauthProvider(providerName)
		if err != nil {
			loginForm(http.StatusConflict, "Your session does not contain the used provider anymore.")
			return
		}

		// finished OAuth flow, exchange code for access token
		accessToken, err := provider.Finish(sess, c.Request)
		sessionCtx.End()

		if err != nil {
			loginForm(http.StatusInternalServerError, err.Error())
			return
		}

		if accessToken == "" {
			loginForm(http.StatusForbidden, "Login cancelled.")
			return
		}

		// fetch the user's profile
		email, _, err := provider.UserProfile(accessToken)
		if err != nil {
			loginForm(http.StatusInternalServerError, err.Error())
			return
		}

		userIdentifier := providerName + ":" + email
		username := config.UsernameByOAuthIdentity(userIdentifier)

		if username == "" {
			loginForm(http.StatusForbidden, "This identity ("+email+") is not assigned to any Hosty account.")
			return
		}

		sess, err = sessionCtx.Start()
		if err != nil {
			loginForm(http.StatusInternalServerError, "Your authentication was successful, but I could not start a session.")
			return
		}

		sess.SetUsername(username)

		target := "/"

		// if the user came from a private/internal file, redirect accordingly
		if dest != "" {
			metadata, err := store.FindByID(dest)
			if err == nil {
				target = fileURI(metadata, false)
			}
		}

		c.Redirect(http.StatusFound, target)
	})
}

func buildOauthProvider(name string) (oauth.Provider, error) {
	info, ok := config.OAuth[name]
	if ok {
		callback := config.Server.BaseUrl + "/oauth/callback"

		switch name {
		case "google":
			return oauth.NewGoogleProvider(info.ClientID, info.ClientSecret, callback, info.Scopes), nil

		case "github":
			return oauth.NewGithubProvider(info.ClientID, info.ClientSecret, callback, info.Scopes), nil
		}
	}

	return nil, errors.New("Invalid provider name given.")
}
