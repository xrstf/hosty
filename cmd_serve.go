package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/xrstf/hosty/session"
)

const (
	sessionName    = "session"
	sessionIDBytes = 32
)

func cmdServe(db *sqlx.DB) {
	gin.SetMode(config.Environment)

	// setup Gin router
	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join(config.Directories.Resources, "templates", "*"))
	router.Static("/assets/", config.Directories.Www)
	router.Use(databaseMiddleware(db))
	router.Use(methodOverride())

	// setup session
	lifetime := config.Session.Lifetime
	if lifetime == nil {
		log.Fatal("No session lifetime configured.\n")
	}

	cookieOpt := session.CookieOptions{
		Name:   config.Session.CookieName,
		MaxAge: *lifetime,
		Secure: config.Session.CookieSecure,
		Path:   config.Session.CookiePath,
	}

	router.Use(session.Middleware(sessionName, *lifetime, cookieOpt, sessionIDBytes))

	setupIndexCtrl(router)
	setupLoginCtrl(router)
	setupOAuthCtrl(router)
	setupPostCtrl(router)
	setupViewCtrl(router)

	// setup our own http server (and configure TLS)
	handler := &maxBytesHandler{
		h: router,
		n: int64(config.Server.MaxRequestSize),
	}

	if config.Server.CertificateFile != "" {
		srv := &http.Server{
			Addr:    config.Server.Listen,
			Handler: handler,
			TLSConfig: &tls.Config{
				CipherSuites: config.CipherSuites(),
			},
		}

		log.Fatal(srv.ListenAndServeTLS(config.Server.CertificateFile, config.Server.PrivateKeyFile))
	} else {
		srv := &http.Server{
			Addr:    config.Server.Listen,
			Handler: handler,
		}

		log.Fatal(srv.ListenAndServe())
	}
}

// based on http://stackoverflow.com/a/28292505/564807
type maxBytesHandler struct {
	h http.Handler
	n int64
}

func (h *maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.n)
	h.h.ServeHTTP(w, r)
}

func requireCsrfToken() gin.HandlerFunc {
	return session.RequireCsrfToken(sessionName)
}
