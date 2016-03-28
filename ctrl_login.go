package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginFormData struct {
	Username    string `form:"username" binding:"required"`
	Password    string `form:"password" binding:"required"`
	Destination string `form:"destination"`
}

type UsersRow struct {
	Login    string
	Password string
}

func setupLoginCtrl(r *gin.Engine) {
	r.GET("/login", func(c *gin.Context) {
		renderLoginForm(c, http.StatusOK, "", "", "")
	})

	r.POST("/login", func(c *gin.Context) {
		// bind form data
		var formData LoginFormData

		form := func(status int, warning string) {
			renderLoginForm(c, status, formData.Destination, formData.Username, warning)
		}

		if c.Bind(&formData) != nil {
			form(http.StatusForbidden, "No username or password given.")
			return
		}

		// find user
		account := config.AccountByUsername(formData.Username)
		if account == nil {
			form(http.StatusForbidden, "Invalid login data given.")
			return
		}

		// check password
		if len(account.Password) == 0 || !compareBcrypt(account.Password, formData.Password) {
			form(http.StatusForbidden, "Invalid login data given.")
			return
		}

		// start session
		ctx := getSessionContext(c)

		session, err := ctx.Start()
		if err != nil {
			form(http.StatusInternalServerError, "Could not start a session: "+err.Error())
			return
		}

		session.SetUsername(formData.Username)

		dest := "/"

		if len(formData.Destination) > 0 {
			metadata, err := store.FindByID(formData.Destination)
			if err == nil {
				dest = fileURI(metadata, false)
			}
		}

		c.Redirect(http.StatusFound, dest)
	})

	r.POST("/logout", requireCsrfToken(), func(c *gin.Context) {
		ctx := getSessionContext(c)
		ctx.End()

		c.Redirect(http.StatusFound, "/")
	})
}

func renderLoginForm(c *gin.Context, status int, destination string, username string, warning string) {
	hasOAuths := len(config.OAuth) > 0
	hasPasswords := false

	for _, acc := range config.Accounts {
		if len(acc.Password) > 0 {
			hasPasswords = true
			break
		}
	}

	_, hasGoogle := config.OAuth["google"]
	_, hasGithub := config.OAuth["github"]

	c.HTML(status, "login.html", gin.H{
		"username":     username,
		"error":        warning,
		"destination":  destination,
		"hasOAuths":    hasOAuths,
		"hasPasswords": hasPasswords,
		"hasGoogle":    hasGoogle,
		"hasGithub":    hasGithub,
	})
}
