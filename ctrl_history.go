package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupHistoryCtrl(r *gin.Engine) {
	r.GET("/history", func(c *gin.Context) {
		ctx := getSessionContext(c)

		if !ctx.IsAuthenticated() {
			renderLoginForm(c, http.StatusUnauthorized, "", "", "")
		} else {
			session := ctx.Session()
			files := store.FindAll(session.Username())
			total := len(files)
			totalSize := 0

			for _, file := range files {
				totalSize = totalSize + file.Size
			}

			viewCtx := getBaseHTMLContext(c)
			viewCtx["files"] = files
			viewCtx["total"] = total
			viewCtx["totalSize"] = totalSize

			c.HTML(http.StatusOK, "history.html", viewCtx)
		}
	})
}
