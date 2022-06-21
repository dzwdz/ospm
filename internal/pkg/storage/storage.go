package storage

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Storage interface {
	List() []string
	Get(path string) ([]byte, error)
	Set(path string, data []byte) error
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

func (s *storage) Get(path string) ([]byte, error) {
	return os.ReadFile(s.base + "/" + filepath.Clean(path))
}

func (s *storage) Set(path string, data []byte) error {
	panic("unimplemented")
}
