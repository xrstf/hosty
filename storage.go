package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xrstf/hosty/session"
)

const (
	selfdestructDisabled = 0
	selfdestructEnabled  = 1
	selfdestructHappened = 2
)

type postMedata struct {
	ID           string  // TEXT NOT NULL PRIMARY KEY,
	Name         string  // TEXT NOT NULL,
	FileType     string  // TEXT NOT NULL,
	Size         int     // INTEGER UNSIGNED NOT NULL,
	Uploaded     int     // INTEGER NOT NULL,
	Uploader     string  // TEXT NOT NULL,
	Expires      *int    // INTEGER NULL,
	Visibility   string  // TEXT NOT NULL
	Selfdestruct int     // INTEGER UNSIGNED NOT NULL,
	Destructor   *string // TEXT NULL
	Legacy       int     // INTEGER UNSIGNED NOT NULL
}

type destructorData struct {
	Time      string
	UserAgent string
	Username  string
	IpAddress string
}

func (pm *postMedata) DestructorData() *destructorData {
	if pm.Destructor == nil {
		return nil
	}

	data := &destructorData{}
	err := json.Unmarshal([]byte(*pm.Destructor), data)
	if err != nil {
		return nil
	}

	return data
}

func NewPostMetadata(name string, uploader string, expiresAt *time.Time, visibility string, selfdestruct bool, filetype string, size int) *postMedata {
	pm := &postMedata{
		Name:         name,
		FileType:     filetype,
		Size:         size,
		Uploaded:     int(time.Now().Unix()),
		Uploader:     uploader,
		Expires:      nil,
		Visibility:   visibility,
		Selfdestruct: 0,
		Destructor:   nil,
		Legacy:       0,
	}

	// self-destructing private files would make no sense
	if visibility != "private" && selfdestruct {
		pm.Selfdestruct = 1
	}

	if expiresAt != nil {
		expires := int(expiresAt.Unix())
		pm.Expires = &expires
	}

	return pm
}

type storage struct {
	database  *sqlx.DB
	directory string
}

func NewStorage(db *sqlx.DB, directory string) *storage {
	store := &storage{db, directory}

	go store.purgeExpired()

	return store
}

func newPostID() (string, error) {
	postID, err := session.RandomString(84) // this is at least twice what we want
	if err != nil {
		return "", errors.New("Could not create post ID: " + err.Error())
	}

	// remove all non-alphanumeric characters
	re := regexp.MustCompile("[^a-zA-Z]")
	postID = re.ReplaceAllString(postID, "")

	// trim down to the actual length we want
	if len(postID) < 42 {
		// this should never happen, unless we had a random string like "abc----------------------------------------------x"
		return newPostID()
	}

	return postID[:42], nil
}

func (s *storage) CreatePaste(content string, highlighted string, metadata *postMedata) (*postMedata, error) {
	if metadata.ID == "" {
		postID, err := newPostID()
		if err != nil {
			return metadata, errors.New("Could not create post ID: " + err.Error())
		}

		metadata.ID = postID
	}

	metadata.Size = len(content)

	err := s.insertIntoDatabase(metadata)
	if err != nil {
		return metadata, err
	}

	dir := s.directory + "/" + metadata.ID

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return metadata, errors.New("Could not create data directory: " + err.Error())
	}

	err = writeFile(dir+"/file.bin", content)
	if err != nil {
		os.RemoveAll(dir)
		return metadata, err
	}

	if highlighted != "" {
		err = writeFile(dir+"/file.html", highlighted)
		if err != nil {
			os.RemoveAll(dir)
			return metadata, err
		}
	}

	return metadata, nil
}

func (s *storage) CreateFile(file io.Reader, metadata *postMedata) (*postMedata, error) {
	if metadata.ID == "" {
		postID, err := newPostID()
		if err != nil {
			return metadata, errors.New("Could not create post ID: " + err.Error())
		}

		metadata.ID = postID
	}

	metadata.Size = 0 // will be set later

	err := s.insertIntoDatabase(metadata)
	if err != nil {
		return metadata, err
	}

	dir := s.directory + "/" + metadata.ID

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return metadata, errors.New("Could not create data directory: " + err.Error())
	}

	err = copyFile(dir+"/file.bin", file)
	if err != nil {
		os.RemoveAll(dir)
		return metadata, err
	}

	stats, err := os.Stat(dir + "/file.bin")
	if err != nil {
		os.RemoveAll(dir)
		return metadata, errors.New("Could not stat data file after uploading it: " + err.Error())
	}

	metadata.Size = int(stats.Size())

	_, err = s.database.Exec(`UPDATE posts SET size = $1 WHERE id = $2`, metadata.Size, metadata.ID)
	if err != nil {
		os.RemoveAll(dir)
		return metadata, errors.New("Could not write to database: " + err.Error())
	}

	return metadata, nil
}

func (s *storage) FindByID(postID string) (*postMedata, error) {
	posts := make([]postMedata, 0)
	s.database.Select(&posts, "SELECT * FROM posts WHERE id = ?", postID)

	if len(posts) != 1 {
		return nil, errors.New("Post not found.")
	}

	// never return expired posts
	post := posts[0]
	now := int(time.Now().Unix())

	if post.Expires != nil && *post.Expires <= now {
		s.Destroy(&post)
		return nil, errors.New("Post not found.")
	}

	return &post, nil
}

func (s *storage) FindRecent(owner string, limit int) []postMedata {
	posts := make([]postMedata, 0)
	s.database.Select(&posts, "SELECT * FROM posts WHERE uploader = $1 ORDER BY uploaded DESC LIMIT $2", owner, limit)

	return posts
}

func (s *storage) Filename(post *postMedata, highlighted bool) (string, error) {
	dir := s.directory + "/" + post.ID

	_, err := os.Stat(dir)
	if err != nil {
		return "", errors.New("Post not found.")
	}

	filename := dir + "/file.bin"

	if highlighted {
		filename = dir + "/file.html"
	}

	_, err = os.Stat(filename)
	if err != nil {
		return "", errors.New("File not found.")
	}

	return filename, nil
}

func (s *storage) Read(post *postMedata, highlighted bool) ([]byte, error) {
	filename, err := s.Filename(post, highlighted)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("Could not read post content from file system.")
	}

	return content, nil
}

func (s *storage) Selfdestruct(post *postMedata, request *http.Request, currentUser string) error {
	if post.Selfdestruct != selfdestructEnabled {
		return errors.New("This post has already been destructed or is not set to destroy itself.")
	}

	// first store some information about this request
	meta := destructorData{
		Time:      time.Now().Format("Mon, 02 Jan 2006 15:04:05 -07:00"),
		UserAgent: request.UserAgent(),
		Username:  currentUser,
		IpAddress: request.RemoteAddr,
	}

	encoded, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	json := string(encoded)

	_, err = s.database.Exec(`UPDATE posts SET destructor = $1, selfdestruct = $2, name = $3 WHERE id = $4`, json, selfdestructHappened, "(deleted)", post.ID)
	if err != nil {
		return errors.New("Could not write to database: " + err.Error())
	}

	// now we can try to remove the files
	err = s.removeFiles(post)
	if err != nil {
		return err
	}

	post.Selfdestruct = selfdestructHappened
	post.Destructor = &json

	return nil
}

func (s *storage) Destroy(post *postMedata) error {
	_, err := s.database.Exec("DELETE FROM posts WHERE id = $1", post.ID)
	if err != nil {
		return errors.New("Could not write to database: " + err.Error())
	}

	dir := s.directory + "/" + post.ID

	_, err = os.Stat(dir)
	if err == nil {
		return os.RemoveAll(dir)
	}

	return nil
}

func (s *storage) removeFiles(post *postMedata) error {
	dir := s.directory + "/" + post.ID

	_, err := os.Stat(dir)
	if err == nil {
		return os.RemoveAll(dir)
	}

	return nil
}

func (s *storage) insertIntoDatabase(post *postMedata) error {
	_, err := s.database.Exec(`
	INSERT INTO posts (id, name, filetype, size, uploaded, uploader, expires, visibility, selfdestruct, legacy)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, post.ID, post.Name, post.FileType, post.Size, post.Uploaded,
		post.Uploader, post.Expires, post.Visibility, post.Selfdestruct,
		post.Legacy)

	if err != nil {
		return errors.New("Could not write to database: " + err.Error())
	}

	return nil
}

func (s *storage) deleteFromDatabase(post *postMedata) error {
	_, err := s.database.Exec("DELETE FROM posts WHERE id = $1", post.ID)
	if err != nil {
		return errors.New("Could not write to database: " + err.Error())
	}

	return nil
}

// purgeExpired is run as a gorutine
func (s *storage) purgeExpired() {
	for {
		// It's not critical to run this often, as the expiry is checked on file
		// access anyway, so this loop is more of a cleanup.
		<-time.After(15 * time.Minute)

		now := int(time.Now().Unix())
		posts := make([]postMedata, 0)
		s.database.Select(&posts, "SELECT * FROM posts WHERE expires IS NOT NULL AND expires < $1 ORDER BY id ASC", now)

		for _, post := range posts {
			err := s.Destroy(&post)
			if err == nil {
				log.Println("Successfully purged expired post " + post.ID + ".")
			} else {
				log.Println("Could not purge post " + post.ID + ": " + err.Error())
			}
		}
	}
}

func writeFile(filename string, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.New("Could not create data file: " + err.Error())
	}

	_, err = file.WriteString(content)
	file.Close()

	if err != nil {
		return errors.New("Could not write to data file: " + err.Error())
	}

	return chmod(filename)
}

func copyFile(dest string, src io.Reader) error {
	file, err := os.Create(dest)
	if err != nil {
		return errors.New("Could not create data file: " + err.Error())
	}

	_, err = io.Copy(file, src)
	file.Close()

	if err != nil {
		return errors.New("Could not write to data file: " + err.Error())
	}

	return chmod(dest)
}

func chmod(filename string) error {
	if err := os.Chmod(filename, 0600); err != nil {
		return errors.New("Could not set permissions on data file: " + err.Error())
	}

	return nil
}
