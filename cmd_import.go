package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

type legacyMetadata struct {
	Filename     string
	Mime         string `json:"type"`
	Language     string
	Size         int
	Selfdestruct bool
	Uploaded     string
	Uploader     string
	AccessTime   string `json:"access_time"`
	AccessBy     string `json:"access_by"`
	AccessUser   string `json:"access_user"`
	AccessFrom   string `json:"access_from"`
}

func cmdImport(db *sqlx.DB) {
	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}

	sourceDir := os.Args[3]
	stats, err := os.Stat(sourceDir)
	if err != nil || !stats.IsDir() {
		log.Fatal("The given directory does not exist or is not a directory.\n")
	}

	nodes, err := filepath.Glob(filepath.Join(sourceDir, "*"))
	if err != nil {
		log.Fatal("Could not glob the directory.\n")
	}

	for _, node := range nodes {
		stats, err := os.Stat(node)
		if err == nil && stats.IsDir() {
			fileID := filepath.Base(node)

			data, err := ioutil.ReadFile(filepath.Join(node, "meta.json"))
			if err != nil {
				log.Printf("[wrn] %s has no meta.json.\n", fileID)
				continue
			}

			meta := legacyMetadata{}
			if err := json.Unmarshal(data, &meta); err != nil {
				log.Printf("[wrn] %s has an invalid meta.json.\n", fileID)
				continue
			}

			uploaded, err := time.Parse("2006-01-02T15:04:05Z", meta.Uploaded)

			if meta.Language != "" {
				// we have a text paste => re-paste it
				content, err := ioutil.ReadFile(filepath.Join(node, "file.bin"))
				if err != nil {
					log.Printf("[wrn] Could not read file.bin from %s (is a paste).\n", fileID)
					continue
				}

				post := NewPostMetadata(
					meta.Filename,
					meta.Uploader,
					nil,
					"public",
					meta.Selfdestruct,
					"text",
					0, // irrelevant, will be set automatically by the storage
				)

				post.ID = fileID
				post.Uploaded = int(uploaded.Unix())
				post.Legacy = 1

				fileIdent := config.FileTypeIdentByPygments(meta.Language)
				highlighted := ""

				if fileIdent != "" {
					fileType := config.FileTypeByIdentifier(fileIdent)

					highlighted, err = highlight(string(content), fileType.Pygments)
					if err != nil {
						log.Printf("[wrn] Could not re-highlight %s.\n", fileID)
						continue
					}
				}

				_, err = store.CreatePaste(string(content), highlighted, post)
				if err != nil {
					log.Printf("[wrn] %s failed: %s.\n", fileID, err.Error())
				} else {
					log.Printf("[inf] %s successfully imported.\n", fileID)
				}
			} else {
				// we have a binary file, i.e. an image
				post := NewPostMetadata(
					meta.Filename,
					meta.Uploader,
					nil,
					"public",
					meta.Selfdestruct,
					config.FileTypeIdentByFilename(meta.Filename),
					0, // irrelevant, will be set automatically by the storage
				)

				post.ID = fileID
				post.Uploaded = int(uploaded.Unix())
				post.Legacy = 1

				f, err := os.Open(filepath.Join(node, "file.bin"))
				if err != nil {
					log.Printf("[wrn] Could not open file.bin from %s.\n", fileID)
					continue
				}

				_, err = store.CreateFile(f, post)
				if err != nil {
					log.Printf("[wrn] %s failed: %s.\n", fileID, err.Error())
				} else {
					log.Printf("[inf] %s successfully imported.\n", fileID)
				}
			}
		}
	}
}
