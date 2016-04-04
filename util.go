package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xrstf/hosty/session"
)

func connectToDatabase(filename string) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	chmod(filename)

	db.MustExec(`CREATE TABLE IF NOT EXISTS posts (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		filetype TEXT NOT NULL,
		size INTEGER UNSIGNED NOT NULL,
		uploaded INTEGER NOT NULL,
		uploader TEXT NOT NULL,
		expires INTEGER NULL,
		visibility TEXT NULL,
		selfdestruct INTEGER UNSIGNED NOT NULL,
		destructor TEXT NULL,
		legacy INTEGER UNSIGNED NOT NULL
	)`)

	return db
}

func databaseMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("database", db)
	}
}

func hashBcrypt(str string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), 12)
	if err != nil {
		panic(err)
	}

	return hash
}

func compareBcrypt(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func highlight(content string, lexer string) (string, error) {
	cmd := exec.Command("pygmentize", "-l", lexer, "-f", "html", "-O", "nowrap,classprefix=pygments-,encoding=utf-8")
	cmd.Stdin = strings.NewReader(content)

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func formatFilesize(num_in uint64) string {
	num := float64(num_in)
	units := []string{"Byte", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"}

	for _, unit := range units {
		if num < 1024.0 {
			// do not show decimals for bytes and kilobytes, because I said so
			if unit == "Byte" || unit == "KiB" {
				return fmt.Sprintf("%.0f %s", num, unit)
			} else {
				return fmt.Sprintf("%.1f %s", num, unit)
			}
		}

		num = (num / 1024)
	}

	return fmt.Sprintf("%.1f %sB", num, "YiB")
}

func fileURI(post *postMedata, raw bool) string {
	uri := "/"

	if raw {
		uri += "r"
	} else {
		uri += "f"
	}

	uri += "/" + post.ID

	// try to append a nice looking name
	cleaned := makeSlug(post.Name, 100)

	if len(cleaned) > 0 {
		uri += "/" + cleaned
	}

	return uri
}

func fileHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha1.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func getSessionContext(c *gin.Context) session.Context {
	return c.MustGet(sessionName).(session.Context)
}

func getSession(c *gin.Context) *session.Session {
	ctx := getSessionContext(c)
	return ctx.Session()
}

func isBlockedUserAgent(req *http.Request) bool {
	ua := req.UserAgent()

	for _, exp := range config.blockedUARegexps {
		if exp.MatchString(ua) {
			return true
		}
	}

	return false
}
