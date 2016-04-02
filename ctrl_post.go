package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PostFormData struct {
	Expire       string `form:"expire" binding:"required"`
	SelfDestruct bool   `form:"selfdestruct"`
	Visibility   string `form:"visibility" binding:"required"`
}

func (p PostFormData) ExpiresAt() (*time.Time, error) {
	expiry := config.Expiry(p.Expire)
	if expiry == nil {
		return nil, errors.New("Invalid expire value given.")
	}

	return expiry.AddTo(time.Now())
}

func (p PostFormData) VisibilityCode() (string, error) {
	switch p.Visibility {
	case "public":
		fallthrough
	case "internal":
		fallthrough
	case "private":
		return p.Visibility, nil
	default:
		return "", errors.New("Invalid visibility value given.")
	}
}

func (p PostFormData) Valid() bool {
	_, err := p.ExpiresAt()
	if err != nil {
		return false
	}

	_, err = p.VisibilityCode()
	if err != nil {
		return false
	}

	return true
}

type PasteFormData struct {
	PostFormData

	Name     string `form:"name" binding:"required"`
	FileType string `form:"filetype"`
	Content  string `form:"content" binding:"required"`
}

func (p PasteFormData) Valid() bool {
	if !p.PostFormData.Valid() {
		return false
	}

	ft := config.FileTypeByIdentifier(p.FileType)
	if ft == nil || len(ft.Pygments) == 0 {
		return false
	}

	return true
}

func setupPostCtrl(r *gin.Engine) {
	r.POST("/paste", requireCsrfToken(), pasteAction)
	r.POST("/upload", requireCsrfToken(), uploadAction)

	// These are actually DELETE requests, but gin seems to not handle them properly
	// (i.e. the default debug log clearly shows that gin has detected a DELETE request,
	// but it won't route accordingly).
	r.POST("/f/:id", requireCsrfToken(), deleteAction)
	r.POST("/f/:id/:trail", requireCsrfToken(), deleteAction)
}

func pasteAction(c *gin.Context) {
	// bind form data
	var formData PasteFormData
	var err error

	errStatus := http.StatusInternalServerError

	if c.Bind(&formData) == nil && formData.Valid() {
		session := getSession(c)
		expiresAt, _ := formData.ExpiresAt()
		visibility, _ := formData.VisibilityCode()
		ft := config.FileTypeByIdentifier(formData.FileType)

		post := NewPostMetadata(
			formData.Name,
			session.Username(),
			expiresAt,
			visibility,
			formData.SelfDestruct,
			formData.FileType,
			0, // irrelevant, will be set automatically by the storage
		)

		err = createPaste(formData.Content, ft.Pygments, post)
		if err == nil {
			c.Redirect(http.StatusFound, fileURI(post, false))
			return
		}
	} else {
		err = errors.New("Invalid form data.")
		errStatus = http.StatusBadRequest
	}

	view := getBaseHTMLContext(c)
	view["name"] = formData.Name
	view["content"] = formData.Content
	view["filetype"] = formData.FileType
	view["warning"] = err.Error()

	c.HTML(errStatus, "index.html", view)
}

func uploadAction(c *gin.Context) {
	// check for a valid file upload
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "No file given.",
		})
		return
	}

	// bind form data
	var formData PostFormData

	if c.Bind(&formData) == nil && formData.Valid() {
		expiresAt, _ := formData.ExpiresAt()
		visibility, _ := formData.VisibilityCode()

		session := getSession(c)

		originalFilename := header.Filename
		fileIdent := config.FileTypeIdentByFilename(originalFilename)
		fileType := config.FileTypeByIdentifier(fileIdent)

		post := NewPostMetadata(
			originalFilename,
			session.Username(),
			expiresAt,
			visibility,
			formData.SelfDestruct,
			fileIdent,
			0, // irrelevant, will be set automatically by the storage
		)

		// if this is a text file, treat it as a paste
		if fileType.DisplayAs == "text" {
			bytes, err := ioutil.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "fail",
					"message": "Could not read request body.",
				})
				return
			}

			err = createPaste(string(bytes), fileType.Pygments, post)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "fail",
					"message": "Could not create paste: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"status": "ok",
				"uri":    fileURI(post, false),
			})
			return
		}

		// we have a binary blob
		_, err = store.CreateFile(file, post)
		if err == nil {
			c.JSON(http.StatusCreated, gin.H{
				"status": "ok",
				"uri":    fileURI(post, false),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": "Could not upload file: " + err.Error(),
			})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "Invalid form data.",
		})
	}
}

func deleteAction(c *gin.Context) {
	if c.PostForm(paramHTTPMethodOverride) != "DELETE" {
		c.String(http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

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

	if currentUser != metadata.Uploader {
		viewCtx["destination"] = metadata.ID
		c.HTML(http.StatusNotFound, "login.html", viewCtx)
		return
	}

	err = store.Destroy(metadata)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func createPaste(content string, language string, post *postMedata) error {
	var err error

	highlighted := ""

	// prevent running Pygments for plain text files, or else uploading large
	// log files would waste CPU cycles on nothing.
	if language != "" && language != "text" {
		highlighted, err = highlight(content, language)
		if err != nil {
			return err
		}
	}

	_, err = store.CreatePaste(content, highlighted, post)

	return err
}
