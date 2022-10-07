package storage

import (
	"errors"
	"io/fs"
	"log"
	"net/url" // sanitizing paths in logs
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

type Storage interface {
	List() []string
	Get(dbPath string) ([]byte, error)
	Add(dbPath string, data []byte) error
}

type storage struct {
	base string
	fs   fs.FS
}

func Init(base string) Storage {
	base, err := filepath.Abs(base)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.ReadDir(base); err != nil { // ensure we can read the directory
		log.Fatal(err)
	}
	return &storage{
		base: base,
		fs:   os.DirFS(base),
	}
}

func (s *storage) List() []string {
	files, err := fs.Glob(s.fs, "**")
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func (s *storage) Get(dbPath string) ([]byte, error) {
	// TODO write path traversal tests
	dbPath = path.Clean(path.Join("/", dbPath))
	log.Printf("db get %s", url.QueryEscape(dbPath))
	return os.ReadFile(path.Join(s.base, dbPath))
}

func (s *storage) Add(dbPath string, data []byte) error {
	match, err := regexp.MatchString("^[A-Za-z0-9 .\\-]+$", dbPath)
	if err != nil || !match {
		// TODO write tests
		if err != nil {
			log.Printf("in storage.Add: %s", err)
		}
		return errors.New("invalid path")
	}
	// TODO would be wise to never overwrite files, or somehow authenticate
	// file creation too
	dbPath = time.Now().Format("2006-01-02") + " " + dbPath
	return os.WriteFile(path.Join(s.base, "/", dbPath), data, 0666)
}
