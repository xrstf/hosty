package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xrstf/hosty/session"
)

type recentUpload struct {
	postMedata

	URI  string
	Date string
}

func setupIndexCtrl(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		ctx := getSessionContext(c)

		if !ctx.IsAuthenticated() {
			renderLoginForm(c, http.StatusUnauthorized, "", "", "")
		} else {
			session := ctx.Session()
			recent := make([]recentUpload, 0)

			for _, upload := range store.FindRecent(session.Username(), 10) {
				recent = append(recent, recentUpload{
					postMedata: upload,
					URI:        fileURI(&upload, false),
					Date:       time.Unix(int64(upload.Uploaded), 0).Format(time.RFC3339),
				})
			}

			viewCtx := getBaseHTMLContext(c)
			viewCtx["recent"] = recent

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

	for _, exp := range config.Expiries {
		expiries = append(expiries, dropdownOption{Key: exp.Ident, Value: exp.Name})
	}

	return gin.H{
		"csrfToken": csrfToken,
		"username":  username,
		"languages": trees,
		"expiries":  expiries,
	}
}
