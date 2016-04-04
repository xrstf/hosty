package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func setupViewCtrl(r *gin.Engine) {
	r.GET("/f/:id", viewFileAction)
	r.GET("/f/:id/:trail", viewFileAction)
	r.GET("/r/:id", rawFileAction)
	r.GET("/r/:id/:trail", rawFileAction)
}

func viewFileAction(c *gin.Context) {
	fileID := c.Param("id")

	viewCtx := getBaseHTMLContext(c)
	viewCtx["fileID"] = fileID

	// find file
	metadata, err := store.FindByID(fileID)
	if err != nil {
		c.HTML(http.StatusNotFound, "not-found.html", viewCtx)
		return
	}

	// check access
	sessionCtx := getSessionContext(c)
	currentUser := sessionCtx.Username()
	isOwner := currentUser == metadata.Uploader
	killItNow := metadata.Selfdestruct == selfdestructEnabled && !isOwner
	stale := metadata.Selfdestruct == selfdestructHappened
	allow := isOwner

	if killItNow && isBlockedUserAgent(c.Request) {
		c.HTML(http.StatusForbidden, "blocked-info.html", viewCtx)
		return
	}

	if stale && !isOwner {
		c.HTML(http.StatusNotFound, "not-found.html", viewCtx)
		return
	}

	switch metadata.Visibility {
	case "public":
		allow = true
	case "internal":
		allow = sessionCtx.IsAuthenticated()
	case "private":
		allow = isOwner
	}

	if !allow {
		renderLoginForm(c, http.StatusUnauthorized, metadata.ID, "", "")
		return
	}

	// if this is a file from Hosty1, we need to handle "/raw" (trail) differently
	if metadata.Legacy == 1 && c.Param("trail") == "raw" {
		c.Redirect(http.StatusMovedPermanently, fileURI(metadata, true))
		return
	}

	content := make([]byte, 0)
	vanished := false

	fileType := config.FileTypeByIdentifier(metadata.FileType)

	if fileType.DisplayAs == "text" {
		content, err = store.Read(metadata, true)

		if err == nil {
			viewCtx["content"] = template.HTML(string(content))
		} else {
			// if we could not read the highlighted version, read the plain version
			content, err = store.Read(metadata, false)
			if err == nil {
				viewCtx["content"] = string(content)
			} else if !stale { // not having any data files is okay when the file is stale
				c.String(http.StatusInternalServerError, "Could not find data file: Database seems to be out of sync.")
				return
			}
		}

		if killItNow {
			if store.Selfdestruct(metadata, c.Request, currentUser) != nil {
				// do not reveal the content if destructing it has failed for some reason
				c.HTML(http.StatusNotFound, "not-found.html", viewCtx)
				return
			}

			vanished = true
		}
	}

	if fileType.IconFile == "" {
		fileType.IconFile = "blank-file"
	}

	iconPath := filepath.Join(config.Directories.Resources, "filetypes", fileType.IconFile+".svg")

	read, err := ioutil.ReadFile(iconPath)
	if err == nil {
		viewCtx["icon"] = template.HTML(string(read))
	} else {
		viewCtx["icon"] = ""
	}

	filehash := ""
	filename, err := store.Filename(metadata, false)
	if err == nil {
		filehash, _ = fileHash(filename)
	}

	viewCtx["file"] = metadata
	viewCtx["isOwner"] = isOwner
	viewCtx["canDelete"] = isOwner
	viewCtx["vanished"] = vanished
	viewCtx["stale"] = stale
	viewCtx["displayType"] = fileType.DisplayAs
	viewCtx["viewURI"] = fileURI(metadata, false)
	viewCtx["rawURI"] = fileURI(metadata, true)
	viewCtx["filehash"] = filehash

	filename = "file-" + fileType.DisplayAs + ".html"

	c.HTML(http.StatusOK, filename, viewCtx)
}

func rawFileAction(c *gin.Context) {
	fileID := c.Param("id")

	// find file
	metadata, err := store.FindByID(fileID)
	if err != nil {
		c.String(http.StatusNotFound, "File not found.")
		return
	}

	// check access
	sessionCtx := getSessionContext(c)
	currentUser := sessionCtx.Username()
	isOwner := currentUser == metadata.Uploader

	// raw access is pointless on stale self-destructive files
	if metadata.Selfdestruct == selfdestructHappened {
		// do not disclose status to non-owners
		if isOwner {
			c.String(http.StatusGone, "File self destructed.")
		} else {
			c.String(http.StatusNotFound, "File not found.")
		}
		return
	}

	// check your privilege
	allow := false

	switch metadata.Visibility {
	case "public":
		allow = true
	case "internal":
		allow = sessionCtx.IsAuthenticated()
	case "private":
		allow = isOwner
	}

	if !allow {
		c.String(http.StatusNotFound, "File not found.")
		return
	}

	filename, err := store.Filename(metadata, false)
	if err != nil {
		c.String(http.StatusNotFound, "File not found.")
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not open data file for reading.")
		return
	}

	if metadata.Selfdestruct == selfdestructEnabled && !isOwner {
		if isBlockedUserAgent(c.Request) {
			c.String(http.StatusForbidden, "Access to this content is not allowed for preview/crawler bots. Please access this link via your browser directly.")
			return
		}

		defer func() {
			store.Selfdestruct(metadata, c.Request, currentUser)
		}()
	}

	// Defer file closing after the selfdestruct because defers are executed
	// in reverse order and we need to close the file handle first.
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not stat() the data file.")
		return
	}

	fileType := config.FileTypeByIdentifier(metadata.FileType)

	// do not rely on Go's content-sniffing
	headers := c.Writer.Header()
	headers.Set("Content-Type", fileType.Mime)
	headers.Set("Content-Disposition", `filename="`+metadata.Name+`"`)

	// prevent caching for non-public files
	if metadata.Visibility != "public" {
		// http://stackoverflow.com/a/2068407/564807
		headers.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")
	}

	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}
