package storage

import (
	"net/url" // sanitizing paths in logs
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

type Storage interface {
	List() []string
	Get(dbPath string) ([]byte, error)
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
