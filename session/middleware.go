package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CookieOptions struct {
	Name   string
	MaxAge time.Duration
	Secure bool
	Path   string
}

func Middleware(sessionName string, ttl time.Duration, opt CookieOptions, idBytes int) gin.HandlerFunc {
	if idBytes <= 0 {
		idBytes = 32
	}

	manager := &manager{
		lock:     sync.RWMutex{},
		sessions: make(map[string]*Session),
		idBytes:  idBytes,
		ttl:      ttl,
		cookie:   opt,
	}

	// cleanup expired sessions
	go func() {
		for {
			<-time.After(15 * time.Minute)
			manager.Cleanup()
		}
	}()

	return func(c *gin.Context) {
		c.Set(sessionName, manager.NewContext(c.Request, c.Writer))
	}
}

type CsrfFormData struct {
	Token string `form:"csrftoken" binding:"required"`
}

func RequireCsrfToken(sessionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet(sessionName).(Context)

		if !ctx.IsAuthenticated() {
			c.String(http.StatusUnauthorized, "Unauthorized.")
			c.Abort()
		} else {
			badToken := true
			session := ctx.Session()

			var formData CsrfFormData

			if c.Bind(&formData) == nil {
				csrfToken := session.CsrfToken()

				if csrfToken == formData.Token {
					badToken = false
				}
			}

			if badToken {
				c.String(http.StatusConflict, "Bad CSRF token.")
				c.Abort()
			}
		}
	}
}

func RandomString(len int) (string, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)

	return base64.RawURLEncoding.EncodeToString(b), err
}
