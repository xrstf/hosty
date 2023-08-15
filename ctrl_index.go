package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.xrstf.de/hosty/session"
)

func setupIndexCtrl(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		ctx := getSessionContext(c)

		if !ctx.IsAuthenticated() {
			renderLoginForm(c, http.StatusUnauthorized, "", "", "")
		} else {
			session := ctx.Session()
			viewCtx := getBaseHTMLContext(c)
			viewCtx["recent"] = store.FindRecent(session.Username(), 10)

			c.HTML(http.StatusOK, "index.html", viewCtx)
		}
	})
}

type dropdownOption struct {
	Key   string
	Value string
}

type languageTree struct {
	GroupName string
	Options   []dropdownOption
}

func getBaseHTMLContext(c *gin.Context) gin.H {
	ctx := c.MustGet("session").(session.Context)
	session := ctx.Session()

	trees := make([]languageTree, 0)

	for _, pastebin := range config.Pastebin {
		t := languageTree{
			GroupName: pastebin.Name,
			Options:   make([]dropdownOption, 0),
		}

		for _, filetype := range pastebin.FileTypes {
			if ft := config.FileTypeByIdentifier(filetype); ft != nil {
				t.Options = append(t.Options, dropdownOption{
					Key:   filetype,
					Value: ft.Name,
				})
			}
		}

		trees = append(trees, t)
	}

	csrfToken := ""
	username := ""

	if session != nil {
		csrfToken = session.CsrfToken()
		username = session.Username()
	}

	expiries := make([]dropdownOption, 0)

	for _, exp := range config.AllowedExpiries(username) {
		expiries = append(expiries, dropdownOption{Key: exp.Ident, Value: exp.Name})
	}

	visibilities := make(map[string]bool)
	canSelfdestruct := false

	for _, vis := range config.AllowedVisibilities(username) {
		visibilities[vis] = true

		if vis == "public" || vis == "internal" {
			canSelfdestruct = true
		}
	}

	hasOptions := len(expiries) > 1 || canSelfdestruct || len(visibilities) > 1

	return gin.H{
		"csrfToken":       csrfToken,
		"username":        username,
		"languages":       trees,
		"expiries":        expiries,
		"visibilities":    visibilities,
		"canSelfdestruct": canSelfdestruct,
		"hasOptions":      hasOptions,
	}
}
